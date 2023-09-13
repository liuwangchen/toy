package csvloader

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"
)

//SettingData 等级表数据
type SettingData struct {
	ID    int32            `json:"id"`
	Value *json.RawMessage `json:"value"`
}

func TestExample(t *testing.T) {

	tt, err := LoadCSVConfig("csvloader_data.txt", reflect.TypeOf(SettingData{}))
	if err != nil {
		panic("load csv error")
	}

	for _, a := range tt {
		_, ok := a.(*SettingData)
		if !ok {
			panic("convert error")
		}
	}
}

func loadCSV() error {
	tData, err := LoadCSVConfig("csvloader_data.txt", reflect.TypeOf(SettingData{}))
	if err != nil {
		return errors.New("load SettingData error")
	}

	typeData := make([]*SettingData, len(tData))
	for i, v := range tData {
		typeV, ok := v.(*SettingData)
		if !ok {
			return errors.New("convert interface{} to struct eror")
		}
		typeData[i] = typeV
	}

	return nil
}
func BenchmarkExample(b *testing.B) {
	for i := 0; i < b.N; i++ {
		loadCSV()
	}
}
