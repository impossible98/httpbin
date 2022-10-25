package info

import (
	// import built-in packages
	"fmt"
	"strconv"
	"time"
	// import local packages
	"httpbin/app/utils/color"
	"httpbin/app/utils/ip"
	"httpbin/global"
)

func ShowInfo(startTime int64) {
	fmt.Print("\033[H\033[2J")
	endTime := time.Now().UnixNano()
	difference := (endTime - startTime) / 1000 / 1000
	// control flow
	if difference == 0 {
		fmt.Printf("  %s %s  %s %s ms\n", color.Bold(color.Green(global.APP_NAME)), color.Green("v"+global.VERSION), color.Faint("ready in"), color.Bold(strconv.Itoa(1)))
	} else {
		fmt.Printf("  %s %s  %s %s ms\n", color.Bold(color.Green(global.APP_NAME)), color.Green("v"+global.VERSION), color.Faint("ready in"), color.Bold(strconv.Itoa(int(difference))))
	}
	fmt.Println()
	fmt.Printf("  %s  %s   %s\n", color.Green("➜"), color.Bold("Local:"), color.Cyan("http://127.0.0.1:"+strconv.Itoa(8080)+"/"))
	fmt.Printf("  %s  %s %s\n", color.Green("➜"), color.Bold("Network:"), color.Cyan("http://"+ip.GetLocalIp()+":"+strconv.Itoa(8080)+"/"))
	fmt.Println()
}
