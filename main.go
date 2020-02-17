package launcher

import (
    "os"
    "fmt"
    "net"
    "time"
    "strconv"
    "runtime"
    "os/user"
    "os/exec"
    "net/http"
    "io/ioutil"
    "path/filepath"
)

var pid string

type Config struct {
    HandleRoot  bool
    Pid     string
    Port    int
}

func Exit() {
    os.Remove(pid)
    go os.Exit(0)
}

func Addr(pid *string) *string {
    if pid == nil || *pid == "" {
        pid = _pid()
    }
    if _, err := os.Stat(*pid); err == nil {
        data, err := ioutil.ReadFile(*pid)
        if err != nil {
            panic(err)
        }
        addr := string(data)
        req, err := http.NewRequest("HEAD", addr, nil)
        if err != nil {
            panic(err)
        }
        client := &http.Client{ Timeout: time.Second * 1 }
        resp, err := client.Do(req)
        if err == nil && resp.StatusCode == 200 {
            return &addr
        }
        os.Remove(*pid)
    }
    return nil
}

func Start() {
    StartWithConfig(nil)
}

func StartWithConfig(c *Config) {

    var port int
    if c != nil {
        port = c.Port
    }

    if c == nil || c.Pid == "" {
        pid = *_pid()
    } else {
        pid = c.Pid
    }
    addr := Addr(&pid)

    if addr != nil && *addr != "" {
        go _open(*addr)
        time.Sleep(time.Second * 1)
        os.Exit(0)
    }

    if c == nil || c.HandleRoot {
        fs := http.FileServer(http.Dir("static/"))
        http.Handle("/", http.StripPrefix("/", fs))
    }

    listener, err := net.Listen("tcp", ":" + strconv.Itoa(port))
    if err != nil {
        panic(err)
    }

    //fmt.Println("Using port:", listener.Addr().(*net.TCPAddr).Port)

    _addr := fmt.Sprintf("%s%d", "http://localhost:", listener.Addr().(*net.TCPAddr).Port)

    if err := os.MkdirAll(filepath.Dir(pid), 0775); err != nil {
        panic(err)
    }
    file, err := os.Create(pid)
    file.WriteString(_addr)
    file.Close()

    go _open(_addr)
    panic(http.Serve(listener, nil))

}

func _pid() *string {
    myself, error := user.Current()
    if error != nil {
        panic(error)
    }
    homedir := myself.HomeDir + "/.hl/"
    pid = homedir + ".pid"
    return &pid
}

// open opens the specified URL in the default browser of the user.
func _open(url string) error {
    var cmd string
    var args []string

    switch runtime.GOOS {
    case "windows":
        cmd = "cmd"
        args = []string{"/c", "start"}
    case "darwin":
        cmd = "open"
    default: // "linux", "freebsd", "openbsd", "netbsd"
        cmd = "xdg-open"
    }
    args = append(args, url)
    return exec.Command(cmd, args...).Start()
}
