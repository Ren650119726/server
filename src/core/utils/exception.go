package utils

import "root/core/log"

func Try(f func()) {
	defer func() {
		if err := recover(); err != nil {
			log.Errorf("Error:%v", err)
		}
	}()
	f()
}
