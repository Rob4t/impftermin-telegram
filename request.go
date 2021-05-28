package main

type (
	appointmentRequestEntry struct {
		AgeIndication            bool     `json:"ageIndication"`
		Appointments             []string `json:"appointments"`
		AutomaticScheduling      int      `json:"automaticScheduling"`
		Birthdate                string   `json:"birthdate"`
		CountryCode              string   `json:"countryCode"`
		City                     string   `json:"city"`
		CustomerStatus           int      `json:"customerStatus"`
		Email                    string   `json:"email"`
		Email2                   string   `json:"email2"`
		FirstCustomer            bool     `json:"firstCustomer"`
		FirstName                string   `json:"firstName"`
		Gender                   string   `json:"gender"`
		IndicationJob            bool     `json:"indicationJob"`
		IndicationMed            bool     `json:"indicationMed"`
		Job                      string   `json:"job"`
		JobIndication            bool     `json:"jobIndication"`
		LastName                 string   `json:"lastName"`
		MedicalIndication        bool     `json:"medicalIndication"`
		Mobilephone              string   `json:"mobilephone"`
		Phone                    string   `json:"phone"`
		SendEmail                bool     `json:"sendEmail"`
		SendLetter               bool     `json:"sendLetter"`
		SendSMS                  bool     `json:"sendSms"`
		StreetName               string   `json:"streetName"`
		StreetNumber             string   `json:"streetNumber"`
		Vaccation                *bool    `json:"vaccation"`
		VaccinationPermit        bool     `json:"-"`
		WaitingListVaccinationPK *int     `json:"waitingListVaccinationPk"`
		Zipcode                  string   `json:"zipcode"`
	}
	appointmentRequest []appointmentRequestEntry
	bookRequestEntry   struct {
		AgeIndication            bool                             `json:"ageIndication"`
		Appointments             []appointmentReserveRequestEntry `json:"appointments"`
		AutomaticScheduling      int                              `json:"automaticScheduling"`
		Birthdate                string                           `json:"birthdate"`
		CountryCode              string                           `json:"countryCode"`
		City                     string                           `json:"city"`
		CustomerStatus           int                              `json:"customerStatus"`
		CustomerCode             string                           `json:"customerCode"`
		CustomerPK               int64                            `json:"customerPk"`
		Email                    string                           `json:"email"`
		Email2                   string                           `json:"email2"`
		FirstCustomer            bool                             `json:"firstCustomer"`
		FirstName                string                           `json:"firstName"`
		Gender                   string                           `json:"gender"`
		Interval1To2             int                              `json:"interval1to2"`
		IndicationJob            bool                             `json:"indicationJob"`
		IndicationMed            bool                             `json:"indicationMed"`
		Job                      string                           `json:"job"`
		JobIndication            bool                             `json:"jobIndication"`
		LastName                 string                           `json:"lastName"`
		MedicalIndication        bool                             `json:"medicalIndication"`
		Mobilephone              string                           `json:"mobilephone"`
		Phone                    string                           `json:"phone"`
		SendEmail                bool                             `json:"sendEmail"`
		SendLetter               bool                             `json:"sendLetter"`
		SendSMS                  bool                             `json:"sendSms"`
		StreetName               string                           `json:"streetName"`
		StreetNumber             string                           `json:"streetNumber"`
		Vaccation                *bool                            `json:"vaccation"`
		VaccinationPermit        bool                             `json:"vaccinationPermit"`
		WaitingListUser          *int                             `json:"waitingListUser"`
		WaitingListVaccinationPK *int                             `json:"waitingListVaccinationPk"`
		Wishlist                 string                           `json:"wishlist"`
		Zipcode                  string                           `json:"zipcode"`
	}
	bookRequest                    []bookRequestEntry
	appointmentReserveRequestEntry struct {
		AppointmentDate     string `json:"appointmentDate"`
		AppointmentPK       int64  `json:"appointmentPk"`
		CustomerPK          int64  `json:"customerPk"`
		Reason              string `json:"reason"`
		VaccinationCenterPK int64  `json:"vaccinationCenterPk"`
		AppointmentStatus   int64  `json:"appointmentStatus"`
		PlatbrixSendStatus  int64  `json:"platbrixSendStatus"`
		DefTime             int64  `json:"defTime"`
		Resend              bool   `json:"resend"`
	}
	appointmentReserveRequest []appointmentReserveRequestEntry
)
