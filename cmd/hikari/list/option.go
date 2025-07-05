package list

type DelegateOption func(d *delegate)

func SetDelegateHeight(v int) DelegateOption {
	return func(d *delegate) {
		d.height = v
	}
}

func SetDelegateSpacing(v int) DelegateOption {
	return func(d *delegate) {
		d.spacing = v
	}
}
