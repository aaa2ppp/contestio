package main

import (
	"log"
	"os"
)

func init() {
	// Устанавливает флаг debug режима, если определена переменная окружения DEBUG.
	//
	// Пример использования (bash) (значение может быть любым и отсутствовать, как в примере):
	//
	//	$ DEBUG= go run .
	_, debug = os.LookupEnv("DEBUG")

	// Дата время нас не интересует, а строчка в коде - важно
	log.SetFlags(log.Llongfile)
}
