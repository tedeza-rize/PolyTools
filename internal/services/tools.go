//go:build windows

package services

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows/registry"
	"polytools/internal/modules"
	"polytools/internal/win32"
)

var hideConsole = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}

// --- tool overlay window (Screen Ruler / Text Extractor) ---

func (s *PolyToolsService) CloseToolWindow() {
	modules.CloseToolWindow()
}

// ExtractRegion captures a screen region, OCRs it via the built-in
// Windows OCR engine (through PowerShell WinRT), and returns the text.
// The result is also copied to the clipboard.
func (s *PolyToolsService) ExtractRegion(x, y, w, h int) (string, error) {
	modules.CloseToolWindow()
	png, err := win32.CaptureRegion(int32(x), int32(y), int32(w), int32(h))
	if err != nil {
		return "", err
	}
	tmp := filepath.Join(os.TempDir(), "polytools-ocr.png")
	if err := os.WriteFile(tmp, png, 0o644); err != nil {
		return "", err
	}
	defer os.Remove(tmp)

	text, err := ocrImage(tmp)
	if err != nil {
		return "", err
	}
	if text != "" {
		win32.SetClipboardText(text)
		win32.Beep()
	}
	return text, nil
}

// ocrImage runs Windows.Media.Ocr through PowerShell on an image file.
func ocrImage(path string) (string, error) {
	script := `
param([string]$Path)
Add-Type -AssemblyName System.Runtime.WindowsRuntime
$null = [Windows.Media.Ocr.OcrEngine, Windows.Foundation, ContentType=WindowsRuntime]
$null = [Windows.Graphics.Imaging.BitmapDecoder, Windows.Foundation, ContentType=WindowsRuntime]
$null = [Windows.Storage.StorageFile, Windows.Storage, ContentType=WindowsRuntime]

function Await($Task, $ResultType) {
  $asTask = [System.WindowsRuntimeSystemExtensions].GetMethod('AsTask', [Type[]]@([Type])).MakeGenericMethod($ResultType)
  $t = $asTask.Invoke($null, @($Task))
  $t.Wait()
  $t.Result
}

$file = Await ([Windows.Storage.StorageFile]::GetFileFromPathAsync($Path)) ([Windows.Storage.StorageFile])
$stream = Await ($file.OpenAsync([Windows.Storage.FileAccessMode]::Read)) ([Windows.Storage.Streams.IRandomAccessStream])
$decoder = Await ([Windows.Graphics.Imaging.BitmapDecoder]::CreateAsync($stream)) ([Windows.Graphics.Imaging.BitmapDecoder])
$bitmap = Await ($decoder.GetSoftwareBitmapAsync()) ([Windows.Graphics.Imaging.SoftwareBitmap])
$engine = [Windows.Media.Ocr.OcrEngine]::TryCreateFromUserProfileLanguages()
$result = Await ($engine.RecognizeAsync($bitmap)) ([Windows.Media.Ocr.OcrResult])
$result.Text
`
	scriptPath := filepath.Join(os.TempDir(), "polytools-ocr.ps1")
	if err := os.WriteFile(scriptPath, []byte(script), 0o644); err != nil {
		return "", err
	}
	defer os.Remove(scriptPath)

	out, err := exec.Command("powershell",
		"-NoProfile", "-STA", "-ExecutionPolicy", "Bypass",
		"-File", scriptPath, "-Path", path).Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// --- Quick Peek ---

type PeekResult struct {
	Path     string `json:"path"`
	Kind     string `json:"kind"`    // text | image | binary | clipboard
	Name     string `json:"name"`
	Text     string `json:"text"`    // text preview or base64 data uri
	Size     int64  `json:"size"`
	Modified string `json:"modified"`
}

var imageExts = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true,
	".bmp": true, ".webp": true, ".ico": true,
}

func (s *PolyToolsService) PeekFile(path string) PeekResult {
	if path == "" {
		return PeekResult{Kind: "clipboard", Text: win32.ClipboardText()}
	}
	info, err := os.Stat(path)
	if err != nil {
		return PeekResult{Kind: "text", Path: path, Name: filepath.Base(path), Text: "(not found)"}
	}
	res := PeekResult{
		Path: path, Name: info.Name(), Size: info.Size(),
		Modified: info.ModTime().Format(time.RFC3339),
	}
	ext := strings.ToLower(filepath.Ext(path))
	raw, err := os.ReadFile(path)
	if err != nil {
		res.Kind, res.Text = "text", "(unreadable: "+err.Error()+")"
		return res
	}
	if imageExts[ext] {
		res.Kind = "image"
		mime := "image/png"
		switch ext {
		case ".jpg", ".jpeg":
			mime = "image/jpeg"
		case ".gif":
			mime = "image/gif"
		case ".webp":
			mime = "image/webp"
		case ".bmp":
			mime = "image/bmp"
		case ".ico":
			mime = "image/x-icon"
		}
		res.Text = "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(raw)
		return res
	}
	res.Kind = "text"
	const maxPeek = 256 * 1024
	if len(raw) > maxPeek {
		raw = raw[:maxPeek]
	}
	// crude binary sniff: NUL byte in first 8KB
	probe := raw
	if len(probe) > 8192 {
		probe = probe[:8192]
	}
	if strings.ContainsRune(string(probe), 0) {
		res.Kind = "binary"
		res.Text = fmt.Sprintf("Binary file — %d bytes", info.Size())
		return res
	}
	res.Text = string(raw)
	return res
}

// --- file pickers for settings panels ---

func (s *PolyToolsService) PickFile(title string) (string, error) {
	return s.app.Dialog.OpenFile().SetTitle(title).PromptForSingleSelection()
}

func (s *PolyToolsService) PickFolder(title string) (string, error) {
	return s.app.Dialog.OpenFile().SetTitle(title).
		CanChooseFiles(false).CanChooseDirectories(true).
		PromptForSingleSelection()
}

// --- Batch Rename ---

type RenamePreview struct {
	Old string `json:"old"`
	New string `json:"new"`
}

type RenameOptions struct {
	Dir       string `json:"dir"`
	Pattern   string `json:"pattern"`
	Replace   string `json:"replace"`
	UseRegex  bool   `json:"useRegex"`
	FilterExt string `json:"filterExt"` // e.g. ".jpg" or "" for all
	Counter   bool   `json:"counter"`   // replace {n} with 1-based counter
}

func (s *PolyToolsService) PreviewRename(o RenameOptions) ([]RenamePreview, error) {
	entries, err := os.ReadDir(o.Dir)
	if err != nil {
		return nil, err
	}
	var re *regexp.Regexp
	if o.UseRegex && o.Pattern != "" {
		re, err = regexp.Compile(o.Pattern)
		if err != nil {
			return nil, fmt.Errorf("invalid regex: %w", err)
		}
	}
	var out []RenamePreview
	n := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if o.FilterExt != "" && !strings.HasSuffix(
			strings.ToLower(name), strings.ToLower(o.FilterExt)) {
			continue
		}
		n++
		var newName string
		if re != nil {
			newName = re.ReplaceAllString(name, o.Replace)
		} else if o.Pattern != "" {
			newName = strings.ReplaceAll(name, o.Pattern, o.Replace)
		} else {
			newName = name
		}
		if o.Counter {
			newName = strings.ReplaceAll(newName, "{n}", fmt.Sprintf("%d", n))
		}
		if newName != name {
			out = append(out, RenamePreview{Old: name, New: newName})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Old < out[j].Old })
	return out, nil
}

func (s *PolyToolsService) ApplyRename(o RenameOptions) (int, error) {
	preview, err := s.PreviewRename(o)
	if err != nil {
		return 0, err
	}
	done := 0
	for _, p := range preview {
		if err := os.Rename(
			filepath.Join(o.Dir, p.Old),
			filepath.Join(o.Dir, p.New),
		); err != nil {
			return done, fmt.Errorf("rename %s: %w", p.Old, err)
		}
		done++
	}
	return done, nil
}

// --- Environment Variables ---

type EnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Scope string `json:"scope"` // user | system
}

const envUserPath = `Environment`
const envSysPath = `SYSTEM\CurrentControlSet\Control\Session Manager\Environment`

func readEnvKey(root registry.Key, path, scope string) []EnvVar {
	k, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
	if err != nil {
		return nil
	}
	defer k.Close()
	names, err := k.ReadValueNames(-1)
	if err != nil {
		return nil
	}
	var out []EnvVar
	for _, n := range names {
		v, _, err := k.GetStringValue(n)
		if err != nil {
			// REG_EXPAND_SZ and others also come back as string via GetStringValue
		}
		out = append(out, EnvVar{Name: n, Value: v, Scope: scope})
	}
	return out
}

func (s *PolyToolsService) EnvVars() []EnvVar {
	out := append(readEnvKey(registry.CURRENT_USER, envUserPath, "user"),
		readEnvKey(registry.LOCAL_MACHINE, envSysPath, "system")...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].Scope != out[j].Scope {
			return out[i].Scope < out[j].Scope
		}
		return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name)
	})
	return out
}

func (s *PolyToolsService) SetEnvVar(name, value, scope string) error {
	root, path := envRoot(scope)
	k, err := registry.OpenKey(root, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if err := k.SetStringValue(name, value); err != nil {
		return err
	}
	broadcastEnvChange()
	return nil
}

func (s *PolyToolsService) DeleteEnvVar(name, scope string) error {
	root, path := envRoot(scope)
	k, err := registry.OpenKey(root, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if err := k.DeleteValue(name); err != nil {
		return err
	}
	broadcastEnvChange()
	return nil
}

func envRoot(scope string) (registry.Key, string) {
	if scope == "system" {
		return registry.LOCAL_MACHINE, envSysPath
	}
	return registry.CURRENT_USER, envUserPath
}

// broadcastEnvChange tells Explorer etc. to reload environment variables.
func broadcastEnvChange() {
	go func() {
		m, _ := syscall.UTF16PtrFromString("Environment")
		user32 := syscall.NewLazyDLL("user32.dll")
		send := user32.NewProc("SendMessageTimeoutW")
		var res uintptr
		send.Call(0xFFFF /* HWND_BROADCAST */, 0x001A /* WM_SETTINGCHANGE */, 0,
			uintptr(unsafe.Pointer(m)), 2 /* SMTO_ABORTIFHUNG */, 5000,
			uintptr(unsafe.Pointer(&res)))
	}()
}

// --- Hosts file ---

const hostsPath = `C:\Windows\System32\drivers\etc\hosts`

func (s *PolyToolsService) ReadHosts() (string, error) {
	raw, err := os.ReadFile(hostsPath)
	return string(raw), err
}

func (s *PolyToolsService) WriteHosts(content string) error {
	err := os.WriteFile(hostsPath, []byte(content), 0o644)
	if err != nil {
		return err
	}
	// flush DNS cache so changes take effect immediately
	c := exec.Command("ipconfig", "/flushdns")
	c.SysProcAttr = hideConsole
	_ = c.Run()
	return nil
}

// --- Automations rules CRUD ---

func (s *PolyToolsService) AutomationRules() []modules.AutomationRule {
	return modules.LoadAutomationRules()
}

func (s *PolyToolsService) SaveAutomationRules(rules []modules.AutomationRule) error {
	if err := modules.SaveAutomationRules(rules); err != nil {
		return err
	}
	// restart the engine so triggers/hotkeys pick up changes
	_ = s.reg.SetEnabled("automations", false)
	_ = s.reg.SetEnabled("automations", true)
	return nil
}

// --- Screensaver ---

func (s *PolyToolsService) DismissScreensaver() {
	modules.ScreensaverDismiss()
}
