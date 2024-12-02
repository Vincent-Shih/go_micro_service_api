package tests

// func TestRegisterRequest(t *testing.T) {
// 	validate := validator.New()

// 	for _, tc := range []struct {
// 		name     string
// 		req      request.RegisterRequest
// 		expected bool
// 	}{
// 		{
// 			name: "success with mobilenumber",
// 			req: request.RegisterRequest{
// 				Account:                "123",
// 				Email:                  "",
// 				CountryCode:            "081",
// 				MobileNumber:           "11111111",
// 				Password:               "1111",
// 				VerificationCodePrefix: "HWP",
// 				VerificationCode:       "123423",
// 				VerificationCodeToken:  "1232321231",
// 			},
// 			expected: true,
// 		},
// 		{
// 			name: "success with email",
// 			req: request.RegisterRequest{
// 				Account:                "123",
// 				Email:                  "test@cus.tw",
// 				CountryCode:            "",
// 				MobileNumber:           "",
// 				Password:               "1111",
// 				VerificationCodePrefix: "HWP",
// 				VerificationCode:       "123423",
// 				VerificationCodeToken:  "1232321231",
// 			},
// 			expected: true,
// 		},
// 		{
// 			name: "empty all",
// 			req: request.RegisterRequest{
// 				Account:                "123",
// 				Email:                  "",
// 				CountryCode:            "",
// 				MobileNumber:           "",
// 				Password:               "1111",
// 				VerificationCodePrefix: "HWP",
// 				VerificationCode:       "123423",
// 				VerificationCodeToken:  "1232321231",
// 			},
// 			expected: false,
// 		},
// 	} {
// 		t.Run(tc.name, func(t *testing.T) {
// 			err := validate.Struct(tc.req)
// 			assert.Equal(t, tc.expected, err == nil)
// 		})
// 	}
// }
