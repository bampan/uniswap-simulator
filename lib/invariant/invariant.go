package invariant

func Invariant(condition bool, message string) {
	if !condition {
		panic(message)
	}
}
