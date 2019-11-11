package jwt

import (
	"fmt"
	"testing"
)

func TestReadFromEnv(t *testing.T) {
	// real app should use longer key than this test
	privKeyPkcs1 := `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----`
	pubKeyPkcs1 := `-----BEGIN RSA PUBLIC KEY-----
MIICCgKCAgEA1t38XB19n7aQi2Jp+SbOOraAlvk1faLaDzJy4RHzftGiElpvleWu
sApoNmoY7Xs0DR+QO+oJKo0Xj71khyKa9ZhbR4TNgcwXnkiLtQL8MbLjIjgDA/v4
lFuE3zyQfm13bDA7abcWe7nukTn7Bh32UIbWPuOHp0DzvLB3nvbiqJD67JmNAuOv
10hZqEcjUB85/xJW/7SbWzcxyDwgbSuApe/P73RqqfXe+xW28ziHmgVKcNXyy1iV
fvjA+HtGoBbTf6S9mz1W3wFfEjiUm+h+niTcSJgun2b5AJqZNH0wlFzV3La7C0U5
GPWEddWmWzNbX/tx9NTeYoAONgNYiBCMnxkqy52jtBh/LYJV90g9I+/SonAMvOid
zY9BPpuxHAvl+km6htbz6+gXdESgC3cS0FbX8Jwt7MWjxpi4i+ExDyqncakYp/fC
uxhHQv0RRlOODDWX444z9PmUnTHgIYWjAQiuYfCQ8qHt2pZEomnGVx4nfMBBGbj4
pgPNc1sDjBomP8IYVhcRqqrNUtRpDSe25TWaXZBvKJ/PiIryInLNqixCcJbf1Jkf
lAonYKes+cwO4vLD6PGk1OhZFA/tkv1lPLPGBkn0NGks4XmQUF33dP5WlRKD/1kM
zpDIXA0lk8M3sCah9pbdl5EVSSNRjjGBHLIiAMxyxhpMiyCXpqGHSU0CAwEAAQ==
-----END RSA PUBLIC KEY-----`
	invalidPubKey := pubKeyPkcs1[:10]
	expireHours := "720"

	jwter, err := NewJWTer(privKeyPkcs1, invalidPubKey, expireHours)
	if err == nil {
		t.Error()
	}

	jwter, err = NewJWTer(privKeyPkcs1, pubKeyPkcs1, expireHours)
	if err != nil {
		t.Error(err)
	}

	type AuthInfo struct {
		UserId   int64
		UserName string
	}
	authInfo := AuthInfo{UserId: 119, UserName: "Đào Thị Lán"}
	token := jwter.CreateAuthToken(authInfo)
	//fmt.Println("token:", token)
	if len(token) == 0 {
		t.Error()
	}

	var auth0 AuthInfo
	err = jwter.CheckAuthToken(token, auth0)
	if err != ErrNonPointerOutput {
		t.Error()
	}

	err = jwter.CheckAuthToken(token, &auth0)
	if err != nil {
		t.Error(err)
	}
	if auth0.UserId != authInfo.UserId ||
		auth0.UserName != authInfo.UserName {
		t.Errorf("error: %#v, %#v", auth0, authInfo)
	}

	type WrongAuthInfo struct {
		UserId   int64
		UserName int64
	}
	var wauth WrongAuthInfo
	err = jwter.CheckAuthToken(token, &wauth)
	//fmt.Println("wauth, err:", wauth, err)
	if err == nil {
		t.Error()
	}
}

func _TestNewJWTer2(t *testing.T) {
	a, b, c := ReadFromEnv()
	_, err := NewJWTer(a, b, c)
	fmt.Println(a, b, c)
	if err != nil {
		t.Error(err)
	}
}
