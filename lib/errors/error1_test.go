package errors

import "fmt"

func top1() error {
	err := middle1()
	if err != nil {
		return Wrap(err, Fields{"topstr": "topstr", "topnum": 1})
	}
	return nil
}

func middle1() error {
	err := bottom1()
	if err != nil {
		num := 1
		return Wrapf(err, Fields{"middlestr": "middlestr", "middlenum": -1}, "middle1 wrapf %d", num).With(Fields{"middleextra": "foo"})
	}
	return nil
}

func bottom1() error {
	return fmt.Errorf("bottom1")
}
