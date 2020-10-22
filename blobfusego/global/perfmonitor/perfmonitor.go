package perfmonitor

type perfStats struct {
	numReadRequest     uint64
	numReadRequestFail uint64
	numReadBytes       uint64

	numWriteRequest     uint64
	numWriteRequestFail uint64
	numWriteBytes       uint64

	numOpenRequest     uint64
	numOpenRequestFail uint64

	numCloseRequest     uint64
	numCloseRequestFail uint64

	numGetAttrRequest     uint64
	numGetAttrRequestFail uint64
}

var currStats = make(map[string]*perfStats)
var prevStats = make(map[string]*perfStats)

func Create(name string) {
	if _, found := currStats[name]; !found {
		currStats[name] = &perfStats{}
		prevStats[name] = &perfStats{}
	}
}

func Delete(name string) {
	if _, found := currStats[name]; !found {
		delete(currStats, name)
		delete(prevStats, name)
	}
}

func ReadRequest(name string) {
	currStats[name].numReadRequest++
}

func ReadRequestFail(name string) {
	currStats[name].numReadRequestFail++
}

func ReadBytes(name string) {
	currStats[name].numReadBytes++
}

func WriteRequest(name string) {
	currStats[name].numWriteRequest++
}

func WriteRequestFail(name string) {
	currStats[name].numWriteRequestFail++
}

func WriteBytes(name string) {
	currStats[name].numWriteBytes++
}

func OpenRequest(name string) {
	currStats[name].numOpenRequest++
}

func OpenRequestFail(name string) {
	currStats[name].numOpenRequestFail++
}

func CloseRequest(name string) {
	currStats[name].numCloseRequest++
}

func CloseRequestFail(name string) {
	currStats[name].numCloseRequestFail++
}

func GetAttrRequest(name string) {
	currStats[name].numGetAttrRequest++
}

func GetAttrRequestFail(name string) {
	currStats[name].numGetAttrRequestFail++
}
