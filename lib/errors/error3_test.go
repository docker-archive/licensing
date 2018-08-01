package errors

func top3() error {
	err := middle3()
	if err != nil {
		return Wrap(err, Fields{"top3": 3})
	}
	return nil
}

func middle3() error {
	err := bottom3()
	if err != nil {
		return Wrapf(err, Fields{"middle3": 3}, "middle3 wrap")
	}
	return nil
}

func bottom3() error {
	return New("rock-bottom")
}
