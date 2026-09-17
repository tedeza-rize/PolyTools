# PolyTools 기능 로드맵 / 설계 노트

각 모듈의 목적, Go+Wails 기준 구현 접근, 난이도, 주의사항.
커밋은 기능(모듈) 단위로 한다.

## 공통 기반 (모듈보다 먼저)

| 작업 | 내용 |
|---|---|
| 핫키 캡처 컨트롤 | 설정 UI에서 단축키 클릭 → 키 조합 직접 입력 → `UpdateSetting("hotkey", ...)` → 재등록. 현재는 읽기 전용 표시 |
| 오버레이 창 헬퍼 | 투명/클릭스루 전체화면 창. Color Picker, Ruler, Find My Mouse, Zone Layouts 등이 공유. Wails 추가 창 or 순수 Win32 layered window (`WS_EX_LAYERED\|WS_EX_TRANSPARENT\|WS_EX_TOPMOST` + `UpdateLayeredWindow`) |
| Win32 래퍼 확장 | `internal/win32`: EnumWindows, GetWindowPlacement, SetWinEventHook, 저수준 훅 공통 설치/해제 (전용 고루틴 + `runtime.LockOSThread` + 메시지 루프) |
| 저수준 훅 유의점 | `WH_KEYBOARD_LL`/`WH_MOUSE_LL` 콜백은 ~300ms(`LowLevelHooksTimeout`) 내 반환 필수. 콜백 안에서는 판정만 하고 실제 작업은 다른 고루틴으로 넘긴다. 관리자 권한 창엔 비관리자 훅이 못 닿음 (UIPI) |

## 모듈별 설계

### ✅ Awake (구현됨)
`SetThreadExecutionState(ES_CONTINUOUS|ES_SYSTEM_REQUIRED[|ES_DISPLAY_REQUIRED])`.
잔여: "N분/시간 동안만" 타이머 모드, 트레이 토글.

### ✅ Always On Top (구현됨)
`RegisterHotKey` → `GetForegroundWindow` + `SetWindowPos(HWND_TOPMOST)` 토글.
잔여: 핀된 창 테두리(overlay), 피드백 사운드, 핫키 커스터마이즈 UI 연동.

### Grab And Move — 난이도 중
- `WH_MOUSE_LL` 훅으로 드래그 감지 + 수식키(기본 Alt) `GetAsyncKeyState` 판정
- 이동: `GetCursorPos` 델타 → `SetWindowPos`; 리사이즈: 우클릭 드래그 또는 가장자리 근접 판정
- 드래그 중 `WM_ENTERSIZEMOVE` 간섭 주의; 스냅된 창은 먼저 restore
- PowerToys의 새 기능이지만 Linux Alt-drag 유저에겐 차별화 포인트

### Window Memory — 난이도 중
- 창 식별: `exe 경로 + window class + title 패턴` (UWP는 ApplicationFrameHost라 `GetApplicationUserModelId` 필요)
- 감지: `SetWinEventHook(EVENT_OBJECT_CREATE/SHOW)` → 열릴 때 복원, `EVENT_OBJECT_LOCATIONCHANGE`/`EVENT_SYSTEM_MOVESIZEEND` → 저장
- 저장 키에 모니터 구성 해시 포함 — 모니터 배치/DPI 바뀌면 무시
- "Windows가 진짜 안 해주는 기능" — 대표 타깃

### Zone Layouts (FancyZones류) — 난이도 상
- 1단계: 프리셋 그리드(2열/3열/4분할) + Shift 드래그 시 존 오버레이 + `EVENT_SYSTEM_MOVESIZEEND`로 드롭 감지 → `SetWindowPos`
- 2단계: 커스텀 레이아웃 에디터(프론트엔드 캔버스), 모니터별 레이아웃
- DPI: manifest `PerMonitorV2` 필요, 스케일 다른 모니터 혼합 테스트 필수

### Text Extractor (OCR) — 난이도 상
- 영역 선택 오버레이 → `Graphics Capture`/`PrintWindow`/`BitBlt` 캡처 → `Windows.Media.Ocr` (WinRT)
- WinRT 호출: `github.com/saltosystems/winrt-go` (코드젠) 검토, 아니면 PowerShell 헬퍼 호출로 시작(느리지만 쉬움)
- 한글 인식은 Windows 설치 OCR 언어팩 의존

### Color Picker — 난이도 하 (다음 후보)
- 전역 핫키 → `GetCursorPos` + `GetDC(0)`/`GetPixel` (또는 1px `BitBlt` 캡처) → 형식 변환 → 클립보드 + 토스트
- 커서 주변 확대 프리뷰 오버레이는 2단계
- 히스토리(최근 색 N개)도 설정 페이지에

### Screen Ruler — 난이도 중
- 오버레이 창에서 드래그 측정; 엣지 감지 모드는 화면 캡처 픽셀 스캔(빛/어둠 경계)
- DPI 스케일링 반영 필수

### Keyboard Manager — 난이도 상
- `WH_KEYBOARD_LL`로 키 인터셉트 → 매핑 테이블 조회 → `SendInput`으로 대체 키 주입 + 원본 억제(`return 1`)
- 단축키 리매핑은 조합 상태 추적 필요. 훅 지연=입력 지연 직결이라 콜백 최소화가 핵심
- UI: 리매핑 테이블 편집기(행 추가/삭제) — 프론트엔드 작업 큼

### Mouse Utilities — 난이도 중
- Find My Mouse: `WH_KEYBOARD_LL`로 Ctrl 2연타 → 커서 외 화면 디밍 오버레이
- Highlighter: `WH_MOUSE_LL` 클릭 시 커서 주변 링 표시
- Crosshair: 전체화면 오버레이에 십자선 (오버레이 헬퍼 재사용)

### Batch Rename — 난이도 중
- 우선 탐색기 통합 없이: 설정 창 내 미니 앱(파일 드롭 → 규칙 → 미리보기 → 실행)
- 탐색기 우클릭 통합: Win11 모던 메뉴는 COM sparse package 필요(큼) — 레거시 `HKCR\*\shell` verb로 우회 검토

### Quick Peek — 난이도 중
- 탐색기 선택 파일: `Shell.Application` COM → `SelectedItems` 경로 획득
- 핫키 → 별도 미리보기 창(텍스트/이미지/메타데이터). 바이너리는 헥스뷰 정도로

### Environment Variables — 난이도 하
- `HKCU\Environment` / `HKLM\SYSTEM\...\Environment` 읽기/쓰기 (`golang.org/x/sys/windows/registry`)
- 변경 후 `SendMessageTimeout(HWND_BROADCAST, WM_SETTINGCHANGE, "Environment")` 브로드캐스트
- 프로필 = 변수 세트 묶어서 적용. 시스템 변수 쓰기엔 관리자 권한 필요

### Hosts File Editor — 난이도 하
- `etc\hosts` 파싱 → 행 목록 UI(활성/비활성 주석 토글) → 저장 시 관리자 권한 필요
- DNS flush 옵션(`ipconfig /flushdns`)

## 제안 순서

1. 공통 기반 (핫키 캡처, 오버레이 헬퍼, 훅 프레임워크)
2. Color Picker / Env Vars / Hosts — 쉬운 실전 모듈로 골격 검증
3. Grab And Move — 훅 프레임워크 첫 적용
4. Window Memory
5. 대형: Zone Layouts → Keyboard Manager → Text Extractor

## 아키텍처 참고

- 모듈이 커지면 `internal/modules/<name>/` 패키지로 분리 (지금 real.go/placeholders.go 분할은 임시)
- 설정 스키마에 `hotkey` 외 타입 추가 가능 (예: `filelist`, `keymap` — Keyboard Manager용)
- 관리자 권한: General에 "관리자로 실행" 토글 → `ShellExecute runas` 재시작 (PowerToys 방식)
