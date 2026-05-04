package confkit

func FromFlags() Source {
	return &flagsSource{}
}

type flagsSource struct{}

func (f *flagsSource) Name() string {
	return "flag"
}

func (f *flagsSource) Lookup(_ *FieldInfo) (any, bool, error) {
	return "", false, nil
}
