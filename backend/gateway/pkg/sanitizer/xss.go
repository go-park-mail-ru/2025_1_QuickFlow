package sanitizer

import (
	"github.com/microcosm-cc/bluemonday"

	"quickflow/gateway/internal/delivery/http/forms"
)

func SanitizeLoginData(loginData *forms.AuthForm, policy *bluemonday.Policy) {
	loginData.Login = policy.Sanitize(loginData.Login)
}

func SanitizeSignUpData(signUpData *forms.SignUpForm, policy *bluemonday.Policy) {
	signUpData.Login = policy.Sanitize(signUpData.Login)
	signUpData.Name = policy.Sanitize(signUpData.Name)
	signUpData.Surname = policy.Sanitize(signUpData.Surname)
	signUpData.DateOfBirth = policy.Sanitize(signUpData.DateOfBirth)
}

func SanitizePost(postData *forms.PostForm, policy *bluemonday.Policy) {
	postData.Text = policy.Sanitize(postData.Text)
}

func SanitizeUpdatePost(postData *forms.UpdatePostForm, policy *bluemonday.Policy) {
	postData.Text = policy.Sanitize(postData.Text)
}

func SanitizeMessage(messageData *forms.MessageForm, policy *bluemonday.Policy) {
	messageData.Text = policy.Sanitize(messageData.Text)
}

func SanitizeProfileInfo(profileData *forms.ProfileInfo, policy *bluemonday.Policy) {
	profileData.Username = policy.Sanitize(profileData.Username)
	profileData.Name = policy.Sanitize(profileData.Name)
	profileData.Surname = policy.Sanitize(profileData.Surname)
	profileData.Bio = policy.Sanitize(profileData.Bio)
	profileData.DateOfBirth = policy.Sanitize(profileData.DateOfBirth)
}

func SanitizeContactInfo(contactInfo *forms.ContactInfo, policy *bluemonday.Policy) {
	contactInfo.Email = policy.Sanitize(contactInfo.Email)
	contactInfo.Phone = policy.Sanitize(contactInfo.Phone)
	contactInfo.City = policy.Sanitize(contactInfo.City)
}

func SanitizeSchoolInfo(schoolInfo *forms.SchoolEducationForm, policy *bluemonday.Policy) {
	schoolInfo.SchoolCity = policy.Sanitize(schoolInfo.SchoolCity)
	schoolInfo.SchoolName = policy.Sanitize(schoolInfo.SchoolName)
}

func SanitizeUniversityInfo(universityInfo *forms.UniversityEducationForm, policy *bluemonday.Policy) {
	universityInfo.UniversityFaculty = policy.Sanitize(universityInfo.UniversityFaculty)
	universityInfo.UniversityCity = policy.Sanitize(universityInfo.UniversityCity)
	universityInfo.UniversityName = policy.Sanitize(universityInfo.UniversityName)
}

func SanitizeFeedbackText(feedback *forms.FeedbackForm, policy *bluemonday.Policy) {
	feedback.Text = policy.Sanitize(feedback.Text)
}

func SanitizeCommunityCreation(community *forms.CreateCommunityForm, policy *bluemonday.Policy) {
	community.Name = policy.Sanitize(community.Name)
	community.Description = policy.Sanitize(community.Description)
}
