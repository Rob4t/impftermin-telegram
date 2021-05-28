package main

type (
	onlySucceededResponse struct {
		Succeeded bool `json:"succeeded"`
	}
	renewTokenResponse struct {
		JWTToken string `json:"jwttoken"`
		Status   string
	}
	availableResponse struct {
		Name                string `json:"name"`
		VaccineName         string `json:"vaccineName"`
		VaccineType         string `json:"vaccineType"`
		OutOfStock          bool   `json:"outOfStock"`
		VaccinationCenterPk int64  `json:"vaccinationCenterPk"`
		Interval1To2        int    `json:"interval1to2"`
	}
	availableListResponse struct {
		ResultList []availableResponse `json:"resultList"`
	}
	appointmentResponse struct {
		CustomerPK       int64  `json:"customerPk"`
		CustomerSequence int64  `json:"customerSequence"`
		CustomerCode     string `json:"customerCode"`
	}
	appointmentListResponse struct {
		ResultList []appointmentResponse `json:"resultList"`
	}
	availableAppointmentsEntryResponse map[string]int
	availableAppointmentsResponse      struct {
		ResultList []availableAppointmentsEntryResponse `json:"resultList"`
	}
	reservedAppointmentsResponse struct {
		ResultList []appointmentReserveRequestEntry `json:"resultList"`
	}
)
