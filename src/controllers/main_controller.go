package controllers

import (
	"github.com/genshen/webConsole/src/models"
	"github.com/genshen/webConsole/src/utils"
	"net/http"
	"strconv"
)

const RunModeProd = "prod"

func Get(w http.ResponseWriter, r *http.Request) {
	// to visit in  vpn mode,please add "vpn" param, e.g. http://console.hpc.gensh.me?vpn=on
	if utils.Config.Site.RunMode == RunModeProd && utils.Config.VPN.Enable && r.URL.Query().Get("vpn") != "" {
		utils.ServeHTTPByName(w, r, "index_vpn.html")
	} else {
		utils.ServeHTTPByName(w, r, "index.html")
	}
}

func SignIn(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method.", 405)
	} else {
		var err error
		var errUnmarshal models.JsonResponse
		err = r.ParseForm()
		if err != nil {
			panic(err)
		}
		userinfo := models.UserInfo{}
		userinfo.Host = r.Form.Get("host")
		port := r.Form.Get("port")
		userinfo.Username = r.Form.Get("username")
		userinfo.Password = r.Form.Get("passwd")

		userinfo.Port, err = strconv.Atoi(port)
		if err != nil {
			userinfo.Port = 22
		}

		if userinfo.Host != "" && userinfo.Username != "" {
			//try to login session account
			session := utils.SSHShellSession{}
			session.Node.Host = userinfo.Host
			session.Node.Port = userinfo.Port
			err := session.Connect(userinfo.Username, userinfo.Password)
			if err != nil {
				errUnmarshal = models.JsonResponse{HasError: true, Message: models.SIGN_IN_FORM_TYPE_ERROR_PASSWORD}
			} else {
				defer session.Close()
				// create session
				client, err := session.GetClient()
				if err != nil {
					// bad connection.
					return
				}
				if session, err := client.NewSession(); err == nil {
					if err := session.Run("whoami"); err == nil {
						if token, expireUnix, err := utils.JwtNewToken(userinfo.Connection, utils.Config.Jwt.Issuer); err == nil {
							errUnmarshal = models.JsonResponse{HasError: false, Addition: token}
							utils.ServeJSON(w, errUnmarshal)
							utils.SessionStorage.Put(token, expireUnix, userinfo)
							return
						}
					}
				}
				errUnmarshal = models.JsonResponse{HasError: true, Message: models.SIGN_IN_FORM_TYPE_ERROR_TEST}
			}
		} else {
			errUnmarshal = models.JsonResponse{HasError: true, Message: models.SIGN_IN_FORM_TYPE_ERROR_VALID}
		}
		utils.ServeJSON(w, errUnmarshal)
	}
}
