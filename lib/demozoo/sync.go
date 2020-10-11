package demozoo

import "errors"

/*
TODO:
Create an automated synchronisation with Demozoo.

For optimisation collect the all existing DZ ids into a searchable array.

Specifically looking to sync the following productions.

Windows

cracktros
https://demozoo.org/productions/?platform=1&production_type=13
MS-DOS

cracktros
https://demozoo.org/productions/?platform=4&production_type=13
bbs
https://demozoo.org/productions/?platform=4&production_type=41
*/

// api = "https://demozoo.org/api/v1/productions"

var ErrSync = errors.New("placeholder")

func Sync() (err error) {
	if ph := 1 + 1; ph == 1 {
		return ErrSync
	}
	return err
}
