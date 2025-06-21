package bus

var blueNormal = []string{"Asrama UI", "Menwa", "Stasiun UI", "Fakultas Psikologi", "FISIP", "Fakultas Ilmu Pengetahuan Budaya", "Fakultas Ekonomi dan Bisnis", "Fakultas Teknik", "Vokasi", "SOR", "FMIPA", "Fakultas Ilmu Keperawatan", "Fakultas Kesehatan Masyarakat", "RIK", "Balairung", "MUI/Perpus UI", "Fakultas Hukum", "Stasiun UI", "Menwa", "Asrama UI", "Parking"}
var blueMorning = []string{"Asrama UI", "Menwa", "Stasiun UI", "Fakultas Psikologi", "FISIP", "Fakultas Ilmu Pengetahuan Budaya", "Fakultas Ekonomi dan Bisnis", "Fakultas Teknik", "Vokasi", "SOR", "FMIPA", "Fakultas Ilmu Keperawatan", "Fakultas Kesehatan Masyarakat", "RIK", "Balairung", "MUI/Perpus UI", "Fakultas Hukum", "Fakultas Psikologi", "FISIP", "Fakultas Ilmu Pengetahuan Budaya", "Fakultas Ekonomi dan Bisnis", "Fakultas Teknik", "Parking"}
var redNormal = []string{"Asrama UI", "Menwa", "Stasiun UI", "Fakultas Hukum", "Balairung", "RIK", "Fakultas Kesehatan Masyarakat", "Fakultas Ilmu Keperawatan", "FMIPA", "SOR", "Vokasi", "Fakultas Teknik", "Fakultas Ekonomi dan Bisnis", "Fakultas Ilmu Pengetahuan Budaya", "FISIP", "Fakultas Psikologi", "Stasiun UI", "Menwa", "Asrama UI", "Parking"}
var redMorning = []string{"Asrama UI", "Menwa", "Stasiun UI", "Fakultas Hukum", "Balairung", "RIK", "Fakultas Kesehatan Masyarakat", "Fakultas Ilmu Keperawatan", "FMIPA", "SOR", "Vokasi", "Fakultas Teknik", "Parking"}

var (
	blueNormalSet  = make(map[[2]string]bool)
	blueMorningSet = make(map[[2]string]bool)
	redNormalSet   = make(map[[2]string]bool)
	redMorningSet  = make(map[[2]string]bool)
)

func init() {
	for i := 0; i < len(blueNormal)-1; i++ {
		blueNormalSet[[2]string{blueNormal[i], blueNormal[i+1]}] = true
	}
	for i := 0; i < len(blueMorning)-1; i++ {
		blueMorningSet[[2]string{blueMorning[i], blueMorning[i+1]}] = true
	}
	for i := 0; i < len(redNormal)-1; i++ {
		redNormalSet[[2]string{redNormal[i], redNormal[i+1]}] = true
	}
	for i := 0; i < len(redMorning)-1; i++ {
		redMorningSet[[2]string{redMorning[i], redMorning[i+1]}] = true
	}
	blueNormalSet[[2]string{blueNormal[len(blueNormal)-1], blueNormal[0]}] = true
	blueMorningSet[[2]string{blueMorning[len(blueMorning)-1], blueMorning[0]}] = true
	redNormalSet[[2]string{redNormal[len(redNormal)-1], redNormal[0]}] = true
	redMorningSet[[2]string{redMorning[len(redMorning)-1], redMorning[0]}] = true
}

func detectRouteColorFromPair(previousHalte, currentHalte string) string {
	if previousHalte == "" || currentHalte == "" {
		return "grey"
	}

	haltePair := [2]string{previousHalte, currentHalte}
	inBlueNormal := blueNormalSet[haltePair]
	inBlueMorning := blueMorningSet[haltePair]
	inRedNormal := redNormalSet[haltePair]
	inRedMorning := redMorningSet[haltePair]
	if (inBlueNormal || inBlueMorning) && !inRedNormal && !inRedMorning {
		return "blue"
	}
	if (inRedNormal || inRedMorning) && !inBlueNormal && !inBlueMorning {
		return "red"
	}
	return "grey"
}
