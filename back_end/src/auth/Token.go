package auth

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"io/ioutil"
	"rabbit_gather/src/logger"
	"rabbit_gather/util"
	"time"
)

var log = logger.NewLoggerWrapper("auth.Token")
var privateKey *rsa.PrivateKey
var publicKey *rsa.PublicKey
var ExpiresAtTimeDuration time.Duration
var NotBeforeTimeDuration time.Duration
var Issuer string
var SignMethod jwt.SigningMethod

const RS256 = "RS256"

func init() {
	type Config struct {
		JwtPrivateKeyFile     string `json:"jwt_private_key_file"`
		JwtPublicKeyFile      string `json:"jwt_public_key_file"`
		ExpiresAtTimeDuration int    `json:"expires_at_time"`
		NotBeforeTimeDuration int    `json:"not_before_time"`
		Issuer                string `json:"issuer"`
		SignMethod            string `json:"sign_method"`
	}

	var config Config
	err := util.ParseJsonConfic(&config, "config/JWT.config.json")
	if err != nil {
		panic(err.Error())
	}

	switch config.SignMethod {
	case RS256:
		SignMethod = jwt.SigningMethodRS256
	default:
		panic("Not supported SignMethod: " + config.SignMethod)
	}

	getPrivateKey := func(theJwtPrivatekeyfile string) (pk *rsa.PrivateKey) {
		privateKeyBytes, err := ioutil.ReadFile(theJwtPrivatekeyfile)
		if err != nil {
			panic(err.Error())
		}
		pk, err = jwt.ParseRSAPrivateKeyFromPEM(privateKeyBytes)
		if err != nil {
			panic(err.Error())
		}
		return
	}
	getPublicKey := func(theTokenPublicKeyFile string) (pk *rsa.PublicKey) {
		publicKeyBytes, err := ioutil.ReadFile(theTokenPublicKeyFile)
		if err != nil {
			panic(err.Error())
		}
		pk, err = jwt.ParseRSAPublicKeyFromPEM(publicKeyBytes)
		if err != nil {
			panic(err.Error())
		}
		return
	}
	ExpiresAtTimeDuration = time.Duration(config.ExpiresAtTimeDuration)
	NotBeforeTimeDuration = time.Duration(config.NotBeforeTimeDuration)
	privateKey = getPrivateKey(config.JwtPrivateKeyFile)
	publicKey = getPublicKey(config.JwtPublicKeyFile)
	Issuer = config.Issuer
}

// Get a new Standard Claims with defult setting.
func NewStandardClaims() *jwt.StandardClaims {
	nowTime := time.Now()
	return &jwt.StandardClaims{
		Audience:  "",
		ExpiresAt: nowTime.Add(ExpiresAtTimeDuration).Unix(),
		Id:        util.Snowflake().String(),
		IssuedAt:  nowTime.Unix(),
		Issuer:    Issuer,
		NotBefore: nowTime.Add(NotBeforeTimeDuration).Unix(),
		Subject:   "",
	}
}

// Parse the jwt token from string and fill in claims.
// Return error when the input JWT token not Vaild
func ParseToken(rawTokenString string, claims UtilityClaims) error {
	var mapClaims jwt.MapClaims //map[string]interface{}//UtilityClaims

	token, err := jwt.ParseWithClaims(rawTokenString, &mapClaims, func(token *jwt.Token) (interface{}, error) {
		e := checkTokenWhenParse(token)
		if e != nil {
			log.DEBUG.Println(e.Error())
		}
		return publicKey, e
	})
	err = claims.ParseMapClaims(&mapClaims)
	if err != nil {
		return err
	}
	//u := map[string]interface{}(mapClaims)

	if err != nil {
		log.DEBUG.Println(err.Error())
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return fmt.Errorf("input string is not a JWT token: %w", err)
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				return fmt.Errorf("token is either expired or not active yet: %w", err)
			} else {
				return fmt.Errorf("this token is not valid: %w", err)
			}
		} else {
			log.ERROR.Println("Error when parse token.")
			return err
		}
	}
	if !token.Valid {
		log.DEBUG.Println(err.Error())

		return fmt.Errorf("this token is not valid: %w", err)
	}
	//mapClaims.GetSubClaims()
	return nil
}

// Check if the token format right.
func checkTokenWhenParse(token *jwt.Token) error {
	if token.Method != SignMethod {
		return errors.New("wrong SignMethod")
	}
	return nil
}

// Sign a new token.
func SignToken(claims *UtilityClaims) (string, error) {
	nowTime := time.Now()
	claims.SetSubClaims(StandardClaimsName, StandardClaims{
		Audience:  "",
		ExpiresAt: nowTime.Add(ExpiresAtTimeDuration).Unix(),
		Id:        util.Snowflake().String(),
		IssuedAt:  nowTime.Unix(),
		Issuer:    Issuer,
		NotBefore: nowTime.Add(NotBeforeTimeDuration).Unix(),
		Subject:   "",
	})
	token := jwt.NewWithClaims(SignMethod, claims)
	signedString, err := token.SignedString(privateKey)
	//jwt.StandardClaims{}
	if err != nil {
		log.ERROR.Println("NewSignedToken Error")
		return "", err
	}
	return signedString, nil
}

func UpdateStatus(existToken string, code StatusBitmask) (string, error) {
	if existToken == "" {
		claims := UtilityClaims{
			StatusClaimsName: StatusClaims{
				StatusBitmask: WaitVerificationCode,
			},
		}
		token, err := SignToken(&claims)
		if err != nil {
			log.DEBUG.Println("fail to create WaitVerificationCode token: ", err.Error())
			return "", err
		}
		return token, nil
	} else {
		//	append permission bit mask
		var uc = UtilityClaims{}
		err := ParseToken(existToken, uc)
		if err != nil {
			log.DEBUG.Println("fail with ParseToken: ", err.Error())
			return "", err
		}

		var statusClaim StatusClaims
		err = uc.MappingClaim(&statusClaim)
		if err != nil {
			log.DEBUG.Println("error when MappingClaim: ", err.Error())
			return "", err
		}
		uc.SetSubClaims(StatusClaimsName, &statusClaim)
		statusClaim.AppendBitMask(WaitVerificationCode)

		token, err := SignToken(&uc)
		if err != nil {
			log.DEBUG.Println("error when Sign token: ", err.Error())
			return "", err
		}
		return token, nil
		//c.JSON(http.StatusOK, gin.H{util.TokenHeaderKey: token})
	}
}

//
//type JWTToken struct {
//	jwt.Token
//	signedString string
//}
//
//func (t *JWTToken) GetSignedString() string {
//	if t.signedString == "" {
//		panic("signedString is empty")
//	}
//	return t.signedString
//}
//
//func (t *JWTToken) AppendBitMask(code StatusBitmask) (*JWTToken, error) {
//	permissionClaims, ok := t.Claims.(*UtilityClaims)
//	if !ok {
//		return nil, errors.New("The Claims is not a UtilityClaims")
//	}
//	if BitMaskCheck(permissionClaims.PermissionBitmask, code) {
//		return t, nil
//	} else {
//		permissionClaims.PermissionBitmask = permissionClaims.PermissionBitmask | code
//	}
//	newToken, err := NewSignedToken(permissionClaims)
//	return newToken, err
//}

// ParseToken Parse the signed token string into claims
//func ParseToken(signedTokenString string, claims jwt.Claims) (*JWTToken, error) {
//if signedTokenString == "" {
//	return nil, errors.New("the token string is \"\"")
//}
//token, err := jwt.ParseWithClaims(signedTokenString, claims, func(token *jwt.Token) (interface{}, error) {
//	e := checkTokenWhenParse(token)
//	return publicKey, e
//})
//if err != nil {
//	return nil, err
//}
//jwtToken := &JWTToken{
//	Token:        *token,
//	signedString: signedTokenString,
//}
//return jwtToken, nil
//}

// NewSignedToken Create new Signed token
//func NewSignedToken(claims jwt.Claims) (*JWTToken, error) {
//	token := jwt.NewWithClaims(SignMethod, claims)
//	signedString, err := token.SignedString(privateKey)
//	if err != nil {
//		log.DEBUG.Println("NewSignedToken Error: ", err.Error())
//		return nil, err
//	}
//	token, err = jwt.ParseWithClaims(signedString, claims, func(token *jwt.Token) (interface{}, error) {
//		return publicKey, nil
//	})
//	if err != nil {
//		log.DEBUG.Println("ParseWithClaims Error: ", err.Error())
//		return nil, err
//	}
//	jwtToken := &JWTToken{
//		Token:        *token,
//		signedString: signedString,
//	}
//	return jwtToken, nil
//}
