package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/bipabo1l/TyrantSocket/protocol"
)

func main() {

	go func() {
		//干活
		for {
			log.Println("++++++++++++干活中++++++++++++++++")
		}

	}()

	go func() {
		ggggggggggggggggg()
	}()

}

//ch1 为-1时，与server端停止通信
var ch1 = make(chan int, 1)
var ch2 = make(chan string, 1)
//ch1 为-1时，与server端停止通信
var chSign = make(chan int, 1)

func sendMsgToServer() {

	go func() {
		//干活
		for {

			select {
			case <-chSign:
				if <-chSign != -1 {

				}
			case <-time.After(time.Second * 3):
				log.Println("++++++++++++干活中++++++++++++++++")
			}

		}

	}()

	go func() {
		ggggggggggggggggg()
	}()

}

func ggggggggggggggggg() {
	//动态传入服务端IP和端口号
	//service := "192.168.0.8:8848"
	service := "127.0.0.1:8848"

	tcpAddr, err := net.ResolveTCPAddr("tcp4", service)

	CheckError(err)

	for {

		conn, err := net.DialTCP("tcp", nil, tcpAddr)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Fatal error:%s", err.Error())
		} else {
			defer conn.Close()

			//连接服务器端成功
			doWork(conn)
		}

		time.Sleep(3 * time.Second)

	}
}

//定义CheckError方法，避免写太多到 if err!=nil
func CheckError(err error) {

	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error:%s", err.Error())

		os.Exit(1)
	}

}

//解决断线重连问题
func doWork(conn net.Conn) error {
	ch1 <- 1
	ch := make(chan int, 100)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case stat := <-ch:
			if stat == 2 {
				return errors.New("None Msg")
			}
		case <-ticker.C:
			ch <- 1
			go ClientMsgHandler(conn, ch)
			go ReadMsg(conn, ch)
		case <-time.After(time.Second * 2):
			defer conn.Close()
			fmt.Println("timeout")
		}

	}

	return nil
}

//客户端消息处理
func ClientMsgHandler(conn net.Conn, ch chan int) {

	<-ch
	//获取当前时间
	//msg := "+++++++++++++++++++++++++++++++++++++"
	SendMsg(conn, "rrrrr")
	go ReadMsg(conn, ch)

}

func GetSession() string {
	gs1 := time.Now().Unix()
	gs2 := strconv.FormatInt(gs1, 10)
	return gs2
}

//接收服务端发来的消息
func ReadMsg(conn net.Conn, ch chan int) {
	<-ch

	//存储被截断的数据
	tmpbuf := make([]byte, 0)
	buf := make([]byte, 1024)

	//将信息解包
	n, _ := conn.Read(buf)
	tmpbuf = protocol.Depack(append(tmpbuf, buf[:n]...))
	msg := string(tmpbuf)
	fmt.Println("server say:", msg)
	if len(msg) == 0 {
		//服务端无返回信息
		ch <- 2
	} else {
		//接收到了服务器端发来的非空包，证明已经和服务器连接成功
		if msg == "RUN" {
			log.Println("收到服务器命令：发送Running包")
			go func() {
				SendMsg(conn, msg)
				time.Sleep(time.Second * 2)
			}()
		} else if msg == "GETSTATUS" {
			//服务器端想知道当前在没在工作
			log.Println("Server端请求了解本Agent状态")
			go SendRunMsg(conn, msg)
		} else if msg == "STOP" {
			//服务器让当前agent停止
			log.Println("服务器让当前agent停止")
			go StopRunMsg(conn, msg)
			//conn.Close()
		} else if msg == "BEGIN" {
			//服务器让当前agent停止
			log.Println("服务器让当前agent重新连接")
			go BeginRunMsg(conn, msg)
			//conn.Close()
		}
	}
}

//向服务端发送消息
func SendMsg(conn net.Conn, msg string) {

	session := GetSession()

	words := "{\"Session\":" + session + ",\"IP\":\"" + get_external() + ",\"Message\":\"" + msg + "\",\"Status\":\"" + "running" + "\"}"
	//将信息封包
	smsg := protocol.Enpack([]byte(words))
	conn.Write(smsg)

}

//向服务端发送消息
func SendRunMsg(conn net.Conn, msg string) {

	session := GetSession()
	words := "{\"Session\":" + session + ",\"IP\":\"" + get_external() + "\",\"Message\":\"" + msg + "\",\"Status\":\"" + "IAMRunning" + "\"}"
	//将信息封包
	smsg := protocol.Enpack([]byte(words))
	conn.Write(smsg)

}

//向服务端发送消息
func StopRunMsg(conn net.Conn, msg string) {

	session := GetSession()
	words := "{\"Session\":" + session + ",\"IP\":\"" + get_external() + "\",\"Message\":\"" + msg + "\",\"Status\":\"" + "STOPPED" + "\"}"
	//将信息封包
	smsg := protocol.Enpack([]byte(words))
	conn.Write(smsg)
	log.Println("----------------------------------")
	//conn.Close()
}

//向服务端发送消息
func BeginRunMsg(conn net.Conn, msg string) {
	session := GetSession()
	words := "{\"Session\":" + session + ",\"IP\":\"" + get_external() + "\",\"Message\":\"" + msg + "\",\"Status\":\"" + "Begin" + "\"}"
	//将信息封包
	smsg := protocol.Enpack([]byte(words))
	conn.Write(smsg)
	log.Println("----------------------------------")
}

func GetMyIP() string {
	addrs, err := net.InterfaceAddrs()
	ip := ""
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, address := range addrs {

		// 检查ip地址判断是否回环地址
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.String()
			}

		}
	}
	return ip
}

func get_external() string {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	content, _ := ioutil.ReadAll(resp.Body)
	log.Println("++++++++++++++++++++++++++++++++++++++++++++")
	log.Println(string(content))
	return string(content)

}
