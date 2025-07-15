package instrumentation

type Config struct{}

func Instrument(_ Config) {
	InitLogger()
	InitTracer()
}
