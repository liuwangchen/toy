package mathx

/*
//获取几分之的几率
func SelectByOdds(upNum, downNum uint32) bool {
	if downNum < 1 {
		return false
	}
	if upNum < 1 {
		return false
	}
	if upNum > downNum-1 {
		return true
	}
	return (1 + uint32((float64(rand.Int63())/(1<<63))*float64(downNum))) <= upNum
}

//获取百分之的几率
func SelectByPercent(percent uint32) bool {
	return SelectByOdds(percent, 100)
}

//获取千分之的几率
func SelectByThousand(th uint32) bool {
	return SelectByOdds(th, 1000)
}

//获取万分之的几率
func SelectByTenTh(tenth uint32) bool {
	return SelectByOdds(tenth, 10000)
}

//获取十万分之的几率
func SelectByLakh(lakh uint32) bool {
	return SelectByOdds(lakh, 100000)
}
*/
