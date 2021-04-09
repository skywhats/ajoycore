package main

import "ajoycore/slog"

func main() {
	slog.Run()
	slog.WriteLog(slog.INFO, "ada", "adadasdadasd")
	slog.Close()
}
