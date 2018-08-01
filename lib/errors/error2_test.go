package errors

func top2() error {
	err := middle2()
	if err != nil {
		return Wrap(err, Fields{"topstr": "topstr", "topnum": 1})
	}
	return nil
}

func middle2() error {
	err := bottom2()
	if err != nil {
		num := 1
		return Wrapf(err, Fields{"middlestr": "middlestr", "middlenum": -1}, "middle1 wrapf %d", num)
	}
	return nil
}

func bottom2() error {
	return NotFound(Fields{"nffield": 1}, "something not found").With(Fields{"other_nffield": 2})
}
