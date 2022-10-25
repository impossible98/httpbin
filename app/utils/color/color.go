package color

func Bold(text string) string {
	return "\x1b[1m" + text + "\x1b[0m"
}

func Faint(text string) string {
	return "\x1b[2m" + text + "\x1b[0m"
}
func Cyan(text string) string {
	return "\x1b[36m" + text + "\x1b[0m"
}

func Green(text string) string {
	return "\x1b[32m" + text + "\x1b[0m"
}
