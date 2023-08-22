package helper

func PanicErr(err error) {
	if err != nil {
		panic(err)
	}
}
