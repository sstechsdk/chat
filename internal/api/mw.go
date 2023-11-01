// Copyright © 2023 OpenIM open source community. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"strconv"

	"github.com/OpenIMSDK/chat/pkg/common/constant"
	constant2 "github.com/OpenIMSDK/protocol/constant"
	"github.com/OpenIMSDK/tools/apiresp"
	"github.com/OpenIMSDK/tools/errs"
	"github.com/OpenIMSDK/tools/log"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"

	"github.com/OpenIMSDK/chat/pkg/proto/admin"
)

func NewMW(adminConn grpc.ClientConnInterface) *MW {
	return &MW{client: admin.NewAdminClient(adminConn)}
}

type MW struct {
	client admin.AdminClient
}

func (o *MW) parseToken(c *gin.Context) (string, int32, string, error) {
	token := c.GetHeader("token")
	if token == "" {
		return "", 0, "", errs.ErrArgs.Wrap("token is empty")
	}
	resp, err := o.client.ParseToken(c, &admin.ParseTokenReq{Token: token})
	if err != nil {
		return "", 0, "", err
	}
	return resp.UserID, resp.UserType, token, nil
}

func (o *MW) parseTokenType(c *gin.Context, userType int32) (string, string, error) {
	userID, t, token, err := o.parseToken(c)
	if err != nil {
		return "", "", err
	}
	if t != userType {
		return "", "", errs.ErrArgs.Wrap("token type error")
	}
	return userID, token, nil
}

func (o *MW) isValidToken(c *gin.Context, userID string, token string) error {
	resp, err := o.client.GetUserToken(c, &admin.GetUserTokenReq{UserID: userID})
	m := resp.TokensMap
	if err != nil {
		log.ZWarn(c, "cache get token error", errs.ErrTokenNotExist.Wrap())
		return err
	}
	if len(m) == 0 {
		log.ZWarn(c, "cache do not exist token error", errs.ErrTokenNotExist.Wrap())
		return errs.ErrTokenNotExist.Wrap()
	}
	if v, ok := m[token]; ok {
		switch v {
		case constant2.NormalToken:
		case constant2.KickedToken:
			log.ZWarn(c, "cache kicked token error", errs.ErrTokenKicked.Wrap())
			return errs.ErrTokenKicked.Wrap()
		default:
			log.ZWarn(c, "cache unknown token error", errs.ErrTokenUnknown.Wrap())
			return err
		}
	} else {
		return errs.ErrTokenNotExist.Wrap()
	}
	return nil
}

func (o *MW) setToken(c *gin.Context, userID string, userType int32) {
	c.Set(constant.RpcOpUserID, userID)
	c.Set(constant.RpcOpUserType, []string{strconv.Itoa(int(userType))})
	c.Set(constant.RpcCustomHeader, []string{constant.RpcOpUserType})
}

func (o *MW) CheckToken(c *gin.Context) {
	userID, userType, token, err := o.parseToken(c)
	if err != nil {
		c.Abort()
		apiresp.GinError(c, err)
		return
	}
	if err := o.isValidToken(c, userID, token); err != nil {
		c.Abort()
		apiresp.GinError(c, err)
		return
	}
	o.setToken(c, userID, userType)
}

func (o *MW) CheckAdmin(c *gin.Context) {
	userID, token, err := o.parseTokenType(c, constant.AdminUser)
	if err != nil {
		c.Abort()
		apiresp.GinError(c, err)
		return
	}
	if err := o.isValidToken(c, userID, token); err != nil {
		c.Abort()
		apiresp.GinError(c, err)
		return
	}
	o.setToken(c, userID, constant.AdminUser)
}

func (o *MW) CheckUser(c *gin.Context) {
	userID, token, err := o.parseTokenType(c, constant.NormalUser)
	if err != nil {
		c.Abort()
		apiresp.GinError(c, err)
		return
	}
	if err := o.isValidToken(c, userID, token); err != nil {
		c.Abort()
		apiresp.GinError(c, err)
		return
	}
	o.setToken(c, userID, constant.NormalUser)
}
