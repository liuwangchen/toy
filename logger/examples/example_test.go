package examples

import (
	"testing"
	"time"

	l4g "github.com/liuwangchen/toy/logger"
)

func TestExample(t *testing.T) {
	err := ExampleYamlFile()
	if err != nil {
		t.Error(err)
		return
	}
}

//func TestLogCompare(te *testing.T) {
//	err := l4g.LoadConfigurationFromFile("example.yaml")
//	if err != nil {
//		return
//	}
//	t := time.Now()
//	for i := 0; i < 100000; i++ {
//		l4g.Info("Info message.")
//	}
//	elapsed := time.Since(t)
//	log.Printf("Elapsed time: %v", elapsed)
//
//	logger, _ := zap.NewProduction()
//	defer logger.Sync()
//
//	t := time.Now()
//	for i := 0; i < 100000; i++ {
//		logger.Info("Info message.")
//	}
//	elapsed := time.Since(t)
//	log.Printf("Elapsed time: %v", elapsed)
//
//	fileLogger := zap.NewExample().Sugar()
//	defer fileLogger.Sync()
//
//	t := time.Now()
//	for i := 0; i < 100000; i++ {
//		fileLogger.Info("Info message.")
//	}
//	elapsed := time.Since(t)
//	log.Printf("Elapsed time: %v", elapsed)
//}

func BenchmarkExample(b *testing.B) {
	// Load the configuration (isn't this easy?)
	err := l4g.LoadConfigurationFromFile("example.yaml")
	if err != nil {
		return
	}
	for i := 0; i < b.N; i++ {
		l4g.Info("About that time, eh chaps? %s", time.Now().String())
	}
	l4g.Close()
}

//func BenchmarkLogger_Logrus(b *testing.B) {
//	logger := logrus.New()
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		logger.Info("Info message.")
//	}
//}

//func BenchmarkLogger_Zap(b *testing.B) {
//	logger, _ := zap.NewProduction()
//	b.ResetTimer()
//
//	for i := 0; i < b.N; i++ {
//		logger.Info("Info message.")
//	}
//}
