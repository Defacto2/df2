package logs_test

import (
	"github.com/Defacto2/df2/pkg/logs"
	"github.com/gookit/color"
)

func ExampleDanger() {
	color.Enable = false
	logs.Panic(false)
	logs.Danger(ErrATest)
	// Output:
}
