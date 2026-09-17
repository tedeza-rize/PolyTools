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

## 신규 제안 기능 검토 (실현 가능성)

### 1. Borderless Gaming — 가능 (중)
- 대상 창에 `SetWindowLongPtr(GWL_STYLE)`로 `WS_OVERLAPPED|WS_CAPTION|WS_THICKFRAME` 등 제거 → 모니터 작업영역 전체로 `SetWindowPos`
- 게임 목록 설정 + `SetWinEventHook(EVENT_OBJECT_CREATE)`으로 창 등장 시 자동 적용
- 한계: exclusive fullscreen(전체화면 전용) 모드는 못 건드림 — 창모드 지원 게임만. 일부 게임이 스타일을 되돌리면 이벤트로 재적용
- 부가: 스타일 복원(토글 해제), 창 위치 지정, 프레임 출력(아래)과 시너지

### 2. 프레임 출력 (FPS 오버레이) — 가능, 우회 경로 (중~상)
- 진짜 인게임 오버레이(Render API 후킹+DLL 인젝션)는 Go로 사실상 불가 — C++ DLL 필요, 안티치트 충돌
- **대안(권장): ETW `Microsoft-Windows-DXGI`/D3D9 Present 이벤트 소비** — PresentMon(MS 오픈소스)이 쓰는 방식. 인젝션 없이 프로세스별 Present 타임스탬프 수집 → FPS/frametime 계산. Go ETW 라이브러리(`bi-zone/etw` 등)로 소비 가능, 안티치트 안전
- 표시: 화면 위 topmost 투명 오버레이 창 — **스타일 설정 가능: 숫자만 / 그래프만 / 숫자+그래프** (카탈로그에 반영됨), 위치 4방향, 프레임타임 토글
- 리스크: Present 이벤트→FPS 해석 로직(PresentMon 참조 필요), 일부 프레젠트 경로 커버리지 차이

### 3. Lossless Scaling — 부분 가능 (상)
- 쉬운 버전(권장 1단계): 창 캡처(`PrintWindow`/Windows Graphics Capture) → 정수배 nearest 스케일 → 보더리스 전체화면 창에 출력 = **integer scaling**. 오래된/저해상도 게임에 실용적
- 고급 버전: GPU 셰이더 업스케일(LS1/FSR1 공개 셰이더) — Go에선 `go-gl`(OpenGL) 컨텍스트 + fragment/compute shader로 이론상 가능하나 공수 큼. Graphics Capture의 WinRT 바인딩(winrt-go) 필요
- 리스크: 캡처→표시 지연시간(인풋랙 체감), 프레임레이트
- 판정: 1단계 정수 스케일링으로 시작, 반응 보고 확장

### 4. 배터리 관리자 (충전 제한 + 알림) — 사용자 구현 진행 중
- **판정 변경: 가능** — 사용자가 Go로 충전 제한을 직접 구현 중 (리눅스 커널 구조 등 활용). 그 구현을 이 모듈에 연결한다
- 설정 스키마(카탈로그 반영됨): `chargeLimitEnabled` + `chargeLimit` 슬라이더(50–100), `notifyEnabled` + `notifyAt` 알림
- 알림 부분은 독립적으로 구현 가능: `GetSystemPowerStatus` 폴링 → 임계치 도달 시 알림/사운드

### 5. 트리거 시스템 (블록 자동화) — 가능 (상이나 핵심 난제는 해결됨)
- **에디터: Google Blockly 임베드** (Scratch/Entry의 실제 엔진, MIT) → 블록을 JSON으로 직렬화 → Go 인터프리터가 트리거별로 실행
- 트리거 후보: 전역 핫키, 시간/스케줄, 창 열림·닫힘·포커스, 프로세스 시작·종료, 파일 변경(fsnotify), 배터리/전원 상태, 클립보드 변경, 유휴 시간
- 액션 후보: 창 조작(이동/크기/최상위/투명), 앱/스크립트 실행, 키·마우스 주입(SendInput 매크로), 볼륨, 알림, HTTP 요청, 파일 작업, 다른 PolyTools 모듈 토글
- 조건 블록: if/else, 비교, 창 타이틀/프로세스명 매치
- 판정: **킬러 차별화 후보** — Windows에 마땅한 비주얼 자동화가 없음(Power Automate Desktop은 무겁고 계정/클라우드 지향). Blockly가 에디터 문제를 해결해 공수는 엔진+트리거 소스 쪽으로 이동
- 단계: (a) JSON 블록 모델+인터프리터+핫키 트리거 (b) Blockly 에디터 UI (c) 트리거/액션 확장

### 6. 단축키(핫키→액션) — 가능 (하~중)
- RegisterHotKey 인프라 이미 있음. "핫키 → 앱 실행/스크립트/키 매크로/모듈 토글" 매핑 리스트
- 5번 트리거 시스템의 트리거 종류로 자연 흡수 가능 — 단독으로 먼저 만들어도 되고
- 조합키 매크로(SendInput 시퀀스)까지 하면 중간

### 7. Win11 우클릭 메뉴 항목 추가 — 부분 가능, 둘 다 지원 (사용자 결정)
- 설정 스키마(카탈로그 반영됨): `classicMenu` / `modernMenu` 토글 — 사용자가 각각 체크
- **클래식 메뉴**: 레거시 verb 등록(`HKCU\Software\Classes\*\shell`, `Directory\shell`, `Directory\Background\shell` 등) → "추가 옵션 표시" 메뉴. 순수 레지스트리 편집 — 쉬움
- **모던 메뉴(최상위)**: sparse MSIX 패키지 + `IExplorerCommand` COM in-proc DLL 필요 → 별도 C++ shim + 패키징. 대형 작업이지만 지원 목표

### 8. 커스텀 화면보호기 (영상/웹) — 가능 (중), 재미있는 차별화
- `.scr` = `/s`(실행) `/c`(설정) `/p`(미리보기) 인자를 처리하는 exe → System32 배치 시 Windows 화면보호기 목록에 등록(관리자 필요)
- `/s` 호출 시 모니터별 전체화면 WebView2 창 → 영상(MP4/WebM/YouTube 임베드), 웹페이지, 이미지 슬라이드, 자체 HTML 등 **웹 기술로 무엇이든**
- 대안 경로: Windows 인프라 안 쓰고 `GetLastInputInfo`로 유휴 감지 → 자체 전체화면 표시 (등록·관리자 불필요, 단 잠금 연동 없음)
- 주의: WebView2 유저데이터 폴더는 쓰기 가능한 경로, 마우스 이동 감지는 초기 위치 스냅샷 후 델타, 멀티모니터 각각 창 생성

## 추가 후보 (Windows 갭 + Go 실현성)

| 기능 | 난이도 | 비고 |
|---|---|---|
| Paste as Plain Text | 하 | 핫키 → 클립보드 서식 제거 → Ctrl+V. 실용적, Windows에 없음 |
| 창 투명화 단축키 | 하 | 임의 창 `SetLayeredWindowAttributes` — 클래식 갭 |
| 창을 트레이로 최소화 | 하~중 | 창 숨김 + 트레이 서브메뉴/아이콘 |
| 모니터 전원 끄기 핫키 | 하 | `WM_SYSCOMMAND SC_MONITORPOWER` — 노트북에 유용 |
| 오디오 장치 전환/앱별 볼륨 | 중 | Core Audio API — `moutend/go-wca` 바인딩 존재 |
| Power Display (DDC/CI) | 중 | 외장 모니터 밝기/대비 — `dxva2.dll` 물리모니터 API |
| 텍스트 확장기 (`;mail`→이메일) | 중 | 키보드 훅 + SendInput. 트리거 시스템 액션으로도 |
| Screen to GIF | 중~상 | 영역 캡처 + `image/gif` 인코딩 = 순수 Go 가능 |
| 화면 감마/색온도 스케줄 | 하 | `SetDeviceGammaRamp` — 나이트라이트 커스텀판 |
| Workspaces (창 배치 저장·복원) | 중 | Window Memory 연계, 수동 트리거형 |
| 입력 시각화 오버레이 | 중 | 스트리머용 키/마우스 표시 — 훅 + 오버레이 헬퍼 |
| 데스크탑 아이콘 토글/정리 | 하 | Progman/SysListView32 show·hide |
| 프로세스 규칙 (우선순위·affinity 고정) | 중 | 프로세스 감시 + SetPriorityClass/SetProcessAffinityMask |
| Wi-Fi/Bluetooth 토글 핫키 | 중 | WinRT Radios API — winrt-go 경유 |

## 제안 순서 (갱신)

1. 공통 기반 (핫키 캡처, 오버레이 헬퍼, 훅 프레임워크)
2. 쉬운 실전: Color Picker, Env Vars, Hosts, Paste as Plain Text, 창 투명화, 충전 알림
3. Grab And Move — 훅 프레임워크 첫 적용
4. Window Memory → Borderless Gaming (창 감시 기반 공유)
5. 트리거 시스템 (킬러 기능 — 핫키 매크로/단축키 기능 흡수)
6. 우클릭 메뉴(클래식), 화면보호기
7. 대형: Zone Layouts → Keyboard Manager → 프레임 출력(ETW) → Text Extractor → Lossless Scaling

## 아키텍처 참고

- 모듈이 커지면 `internal/modules/<name>/` 패키지로 분리 (지금 real.go/placeholders.go 분할은 임시)
- 설정 스키마에 `hotkey` 외 타입 추가 가능 (예: `filelist`, `keymap` — Keyboard Manager용)
- 관리자 권한: General에 "관리자로 실행" 토글 → `ShellExecute runas` 재시작 (PowerToys 방식)
