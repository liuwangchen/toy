package examples

import l4g "github.com/liuwangchen/toy/logger"

func ExampleYamlFile() error {
	// Load the configuration (isn't this easy?)
	err := l4g.LoadConfigurationFromFile("example.yaml")
	if err != nil {
		return err
	}

	// And now we're ready!
	l4g.Finest("This will only go to those of you really cool UDP kids!  If you change enabled=true.")
	l4g.Debug("Oh no!  %d + %d = %d!", 2, 2, 2+2)
	l4g.Info("About that time, eh chaps?")
	l4g.Close()
	return nil
}
