package common

type ErrStruct struct {
	Message string `json:"message"`
}

func Check(err error) {
	if err != nil {
		panic(err)
	}
}