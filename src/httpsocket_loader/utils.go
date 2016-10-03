package main

func dieOnError(err error) {
	if (err != nil) {
		panic(err)
	}
}
