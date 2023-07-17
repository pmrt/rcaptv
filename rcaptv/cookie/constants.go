package cookie

type ClaimKey string

const (
	CredentialsCookie = "credentials"
	UserCookie        = "user"
	OauthStateCookie  = "oauth_state"

	AccessToken  ClaimKey = "a"
	RefreshToken ClaimKey = "r"
	Expiry       ClaimKey = "e"
	UserId       ClaimKey = "u"

	TwitchID       ClaimKey = "twitch_id"
	Username       ClaimKey = "login"
	DisplayName    ClaimKey = "display_name"
	ProfilePicture ClaimKey = "profile_picture"
	BcType         ClaimKey = "bc_type"
)
