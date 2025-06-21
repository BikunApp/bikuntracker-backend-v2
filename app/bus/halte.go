package bus

import "math"

type Halte struct {
	Name     string
	Lat, Lng float64
}

var halteList = []Halte{
	{"Asrama UI", -6.348351370044594, 106.82976588606834},
	{"Menwa", -6.353471269466313, 106.83177955448627},
	{"Stasiun UI", -6.361052900888018, 106.83170076459645},
	{"Fakultas Psikologi", -6.36255935735158, 106.83111906051636},
	{"FISIP", -6.361574, 106.830172},
	{"Fakultas Ilmu Pengetahuan Budaya", -6.361254501381427, 106.82978868484497},
	{"Fakultas Ekonomi dan Bisnis", -6.35946048561971, 106.82582974433899},
	{"Fakultas Teknik", -6.361043911445512, 106.82325214147568},
	{"Vokasi", -6.366036735678631, 106.8216535449028},
	{"SOR", -6.366915739619239, 106.82448193430899},
	{"FMIPA", -6.369828304090281, 106.8257811293006},
	{"Fakultas Ilmu Keperawatan", -6.371008186217929, 106.8268945813179},
	{"Fakultas Kesehatan Masyarakat", -6.371677262480034, 106.8293622136116},
	{"RIK", -6.36987795182555, 106.8310546875},
	{"Balairung", -6.368212251024606, 106.83178257197142},
	{"MUI/Perpus UI", -6.3655942342627565, 106.83204710483551},
	{"Fakultas Hukum", -6.364901492199248, 106.83221206068993},
	{"Parking", -6.348922, 106.826476},
}

func nearestHalte(lat, lng float64) (string, float64) {
	const earthRadius = 6371000 // meters
	minDist := 1e9
	closest := ""
	for _, halte := range halteList {
		dLat := (halte.Lat - lat) * (3.141592653589793 / 180)
		dLng := (halte.Lng - lng) * (3.141592653589793 / 180)
		alat := lat * (3.141592653589793 / 180)
		blat := halte.Lat * (3.141592653589793 / 180)
		a := (dLat/2)*(dLat/2) + (dLng/2)*(dLng/2)*cos(alat)*cos(blat)
		c := 2 * atan2Sqrt(a, 1-a)
		dist := earthRadius * c
		if dist < minDist {
			minDist = dist
			closest = halte.Name
		}
	}
	return closest, minDist
}

func cos(x float64) float64 {
	return math.Cos(x)
}
func atan2Sqrt(a, b float64) float64 {
	return math.Atan2(math.Sqrt(a), math.Sqrt(b))
}
