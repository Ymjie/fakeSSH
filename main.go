package main

import (
	"flag"
	"fmt"
	"github.com/gliderlabs/ssh"
	gossh "golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

var (
	port  int
	local string
)

func main() {
	hostKeySigner, err := createOrLoadKeySigner()
	if err != nil {
		log.Fatal(err)
	}
	flag.StringVar(&local, "local", "0.0.0.0", "local listening port")
	flag.IntVar(&port, "p", 22, "SSH Server Port")
	flag.Parse()
	s := &ssh.Server{
		Addr:            fmt.Sprintf("%v:%v", local, port),
		Handler:         SSHHandler,
		PasswordHandler: passwordHandler,
	}
	s.AddHostKey(hostKeySigner)
	log.Fatal(s.ListenAndServe())
}

func passwordHandler(ctx ssh.Context, password string) bool {
	data := fmt.Sprintf("[%s]<  %s:%s  >from:%s\n", time.Now().Format("2006-01-02 15:04:05"), ctx.User(), password, ctx.RemoteAddr())
	fmt.Print(data)
	file, err := os.OpenFile("pw.txt", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}
	defer file.Close()
	file.Write([]byte(data))
	return false
}
func SSHHandler(s ssh.Session) {
	s.Write([]byte("log"))
	s.Exit(1)

}

//创建key 来验证 host public
func createOrLoadKeySigner() (gossh.Signer, error) {
	keyPath := filepath.Join(os.TempDir(), "fssh.rsa")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(keyPath), os.ModePerm)
		stderr, err := exec.Command("ssh-keygen", "-f", keyPath, "-t", "rsa", "-N", "").CombinedOutput()
		output := string(stderr)
		if err != nil {
			return nil, fmt.Errorf("Fail to generate private key: %v - %s", err, output)
		}
	}
	privateBytes, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}
	return gossh.ParsePrivateKey(privateBytes)
}
