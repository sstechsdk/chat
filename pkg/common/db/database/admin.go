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

package database

import (
	"context"

	"github.com/OpenIMSDK/chat/pkg/common/db/cache"
	"github.com/OpenIMSDK/protocol/constant"
	"github.com/redis/go-redis/v9"

	"github.com/OpenIMSDK/tools/tx"
	"gorm.io/gorm"

	"github.com/OpenIMSDK/chat/pkg/common/db/model/admin"
	table "github.com/OpenIMSDK/chat/pkg/common/db/table/admin"
)

type AdminDatabaseInterface interface {
	InitAdmin(ctx context.Context) error
	GetAdmin(ctx context.Context, account string) (*table.Admin, error)
	GetAdminUserID(ctx context.Context, userID string) (*table.Admin, error)
	UpdateAdmin(ctx context.Context, userID string, update map[string]any) error
	CreateApplet(ctx context.Context, applets ...*table.Applet) error
	DelApplet(ctx context.Context, appletIDs []string) error
	GetApplet(ctx context.Context, appletID string) (*table.Applet, error)
	FindApplet(ctx context.Context, appletIDs []string) ([]*table.Applet, error)
	SearchApplet(ctx context.Context, keyword string, page int32, size int32) (uint32, []*table.Applet, error)
	FindOnShelf(ctx context.Context) ([]*table.Applet, error)
	UpdateApplet(ctx context.Context, appletID string, update map[string]any) error
	GetConfig(ctx context.Context) (map[string]string, error)
	SetConfig(ctx context.Context, cs map[string]string) error
	DelConfig(ctx context.Context, keys []string) error
	FindInvitationRegister(ctx context.Context, codes []string) ([]*table.InvitationRegister, error)
	DelInvitationRegister(ctx context.Context, codes []string) error
	UpdateInvitationRegister(ctx context.Context, code string, fields map[string]any) error
	CreatInvitationRegister(ctx context.Context, invitationRegisters []*table.InvitationRegister) error
	SearchInvitationRegister(ctx context.Context, keyword string, state int32, userIDs []string, codes []string, page int32, size int32) (uint32, []*table.InvitationRegister, error)
	SearchIPForbidden(ctx context.Context, keyword string, state int32, page int32, size int32) (uint32, []*table.IPForbidden, error)
	AddIPForbidden(ctx context.Context, ms []*table.IPForbidden) error
	FindIPForbidden(ctx context.Context, ms []string) ([]*table.IPForbidden, error)
	DelIPForbidden(ctx context.Context, ips []string) error
	FindDefaultFriend(ctx context.Context, userIDs []string) ([]string, error)
	AddDefaultFriend(ctx context.Context, ms []*table.RegisterAddFriend) error
	DelDefaultFriend(ctx context.Context, userIDs []string) error
	SearchDefaultFriend(ctx context.Context, keyword string, page int32, size int32) (uint32, []*table.RegisterAddFriend, error)
	FindDefaultGroup(ctx context.Context, groupIDs []string) ([]string, error)
	AddDefaultGroup(ctx context.Context, ms []*table.RegisterAddGroup) error
	DelDefaultGroup(ctx context.Context, groupIDs []string) error
	SearchDefaultGroup(ctx context.Context, keyword string, page int32, size int32) (uint32, []*table.RegisterAddGroup, error)
	FindBlockInfo(ctx context.Context, userIDs []string) ([]*table.ForbiddenAccount, error)
	GetBlockInfo(ctx context.Context, userID string) (*table.ForbiddenAccount, error)
	BlockUser(ctx context.Context, f []*table.ForbiddenAccount) error
	DelBlockUser(ctx context.Context, userID []string) error
	SearchBlockUser(ctx context.Context, keyword string, page int32, size int32) (uint32, []*table.ForbiddenAccount, error)
	FindBlockUser(ctx context.Context, userIDs []string) ([]*table.ForbiddenAccount, error)
	SearchUserLimitLogin(ctx context.Context, keyword string, page int32, size int32) (uint32, []*table.LimitUserLoginIP, error)
	AddUserLimitLogin(ctx context.Context, ms []*table.LimitUserLoginIP) error
	DelUserLimitLogin(ctx context.Context, ms []*table.LimitUserLoginIP) error
	CountLimitUserLoginIP(ctx context.Context, userID string) (uint32, error)
	GetLimitUserLoginIP(ctx context.Context, userID string, ip string) (*table.LimitUserLoginIP, error)
	CacheToken(ctx context.Context, userID string, token string) error
	GetTokens(ctx context.Context, userID string) (map[string]int32, error)
}

func NewAdminDatabase(db *gorm.DB, rdb redis.UniversalClient) AdminDatabaseInterface {
	return &AdminDatabase{
		tx:                 tx.NewGorm(db),
		admin:              admin.NewAdmin(db),
		ipForbidden:        admin.NewIPForbidden(db),
		forbiddenAccount:   admin.NewForbiddenAccount(db),
		limitUserLoginIP:   admin.NewLimitUserLoginIP(db),
		invitationRegister: admin.NewInvitationRegister(db),
		registerAddFriend:  admin.NewRegisterAddFriend(db),
		registerAddGroup:   admin.NewRegisterAddGroup(db),
		applet:             admin.NewApplet(db),
		clientConfig:       admin.NewClientConfig(db),
		cache:              cache.NewTokenInterface(rdb),
	}
}

type AdminDatabase struct {
	tx                 tx.Tx
	admin              table.AdminInterface
	ipForbidden        table.IPForbiddenInterface
	forbiddenAccount   table.ForbiddenAccountInterface
	limitUserLoginIP   table.LimitUserLoginIPInterface
	invitationRegister table.InvitationRegisterInterface
	registerAddFriend  table.RegisterAddFriendInterface
	registerAddGroup   table.RegisterAddGroupInterface
	applet             table.AppletInterface
	clientConfig       table.ClientConfigInterface
	cache              cache.TokenInterface
}

func (o *AdminDatabase) InitAdmin(ctx context.Context) error {
	return o.admin.InitAdmin(ctx)
}

func (o *AdminDatabase) GetAdmin(ctx context.Context, account string) (*table.Admin, error) {
	return o.admin.Take(ctx, account)
}

func (o *AdminDatabase) GetAdminUserID(ctx context.Context, userID string) (*table.Admin, error) {
	return o.admin.TakeUserID(ctx, userID)
}

func (o *AdminDatabase) UpdateAdmin(ctx context.Context, userID string, update map[string]any) error {
	return o.admin.Update(ctx, userID, update)
}

func (o *AdminDatabase) CreateApplet(ctx context.Context, applets ...*table.Applet) error {
	return o.applet.Create(ctx, applets...)
}

func (o *AdminDatabase) DelApplet(ctx context.Context, appletIDs []string) error {
	return o.applet.Del(ctx, appletIDs)
}

func (o *AdminDatabase) GetApplet(ctx context.Context, appletID string) (*table.Applet, error) {
	return o.applet.Take(ctx, appletID)
}

func (o *AdminDatabase) FindApplet(ctx context.Context, appletIDs []string) ([]*table.Applet, error) {
	return o.applet.FindID(ctx, appletIDs)
}

func (o *AdminDatabase) SearchApplet(ctx context.Context, keyword string, page int32, size int32) (uint32, []*table.Applet, error) {
	return o.applet.Search(ctx, keyword, page, size)
}

func (o *AdminDatabase) FindOnShelf(ctx context.Context) ([]*table.Applet, error) {
	return o.applet.FindOnShelf(ctx)
}

func (o *AdminDatabase) UpdateApplet(ctx context.Context, appletID string, update map[string]any) error {
	return o.applet.Update(ctx, appletID, update)
}

func (o *AdminDatabase) GetConfig(ctx context.Context) (map[string]string, error) {
	return o.clientConfig.Get(ctx)
}

func (o *AdminDatabase) SetConfig(ctx context.Context, cs map[string]string) error {
	return o.clientConfig.Set(ctx, cs)
}

func (o *AdminDatabase) DelConfig(ctx context.Context, keys []string) error {
	return o.clientConfig.Del(ctx, keys)
}

func (o *AdminDatabase) FindInvitationRegister(ctx context.Context, codes []string) ([]*table.InvitationRegister, error) {
	return o.invitationRegister.Find(ctx, codes)
}

func (o *AdminDatabase) DelInvitationRegister(ctx context.Context, codes []string) error {
	return o.invitationRegister.Del(ctx, codes)
}

func (o *AdminDatabase) UpdateInvitationRegister(ctx context.Context, code string, fields map[string]any) error {
	return o.invitationRegister.Update(ctx, code, fields)
}

func (o *AdminDatabase) CreatInvitationRegister(ctx context.Context, invitationRegisters []*table.InvitationRegister) error {
	return o.invitationRegister.Create(ctx, invitationRegisters...)
}

func (o *AdminDatabase) SearchInvitationRegister(ctx context.Context, keyword string, state int32, userIDs []string, codes []string, page int32, size int32) (uint32, []*table.InvitationRegister, error) {
	return o.invitationRegister.Search(ctx, keyword, state, userIDs, codes, page, size)
}

func (o *AdminDatabase) SearchIPForbidden(ctx context.Context, keyword string, state int32, page int32, size int32) (uint32, []*table.IPForbidden, error) {
	return o.ipForbidden.Search(ctx, keyword, state, page, size)
}

func (o *AdminDatabase) AddIPForbidden(ctx context.Context, ms []*table.IPForbidden) error {
	return o.ipForbidden.Create(ctx, ms)
}

func (o *AdminDatabase) FindIPForbidden(ctx context.Context, ms []string) ([]*table.IPForbidden, error) {
	return o.ipForbidden.Find(ctx, ms)
}

func (o *AdminDatabase) DelIPForbidden(ctx context.Context, ips []string) error {
	return o.ipForbidden.Delete(ctx, ips)
}

func (o *AdminDatabase) FindDefaultFriend(ctx context.Context, userIDs []string) ([]string, error) {
	return o.registerAddFriend.FindUserID(ctx, userIDs)
}

func (o *AdminDatabase) AddDefaultFriend(ctx context.Context, ms []*table.RegisterAddFriend) error {
	return o.registerAddFriend.Add(ctx, ms)
}

func (o *AdminDatabase) DelDefaultFriend(ctx context.Context, userIDs []string) error {
	return o.registerAddFriend.Del(ctx, userIDs)
}

func (o *AdminDatabase) SearchDefaultFriend(ctx context.Context, keyword string, page int32, size int32) (uint32, []*table.RegisterAddFriend, error) {
	return o.registerAddFriend.Search(ctx, keyword, page, size)
}

func (o *AdminDatabase) FindDefaultGroup(ctx context.Context, groupIDs []string) ([]string, error) {
	return o.registerAddGroup.FindGroupID(ctx, groupIDs)
}

func (o *AdminDatabase) AddDefaultGroup(ctx context.Context, ms []*table.RegisterAddGroup) error {
	return o.registerAddGroup.Add(ctx, ms)
}

func (o *AdminDatabase) DelDefaultGroup(ctx context.Context, groupIDs []string) error {
	return o.registerAddGroup.Del(ctx, groupIDs)
}

func (o *AdminDatabase) SearchDefaultGroup(ctx context.Context, keyword string, page int32, size int32) (uint32, []*table.RegisterAddGroup, error) {
	return o.registerAddGroup.Search(ctx, keyword, page, size)
}

func (o *AdminDatabase) FindBlockInfo(ctx context.Context, userIDs []string) ([]*table.ForbiddenAccount, error) {
	return o.forbiddenAccount.Find(ctx, userIDs)
}

func (o *AdminDatabase) GetBlockInfo(ctx context.Context, userID string) (*table.ForbiddenAccount, error) {
	return o.forbiddenAccount.Take(ctx, userID)
}

func (o *AdminDatabase) BlockUser(ctx context.Context, f []*table.ForbiddenAccount) error {
	return o.forbiddenAccount.Create(ctx, f)
}

func (o *AdminDatabase) DelBlockUser(ctx context.Context, userID []string) error {
	return o.forbiddenAccount.Delete(ctx, userID)
}

func (o *AdminDatabase) SearchBlockUser(ctx context.Context, keyword string, page int32, size int32) (uint32, []*table.ForbiddenAccount, error) {
	return o.forbiddenAccount.Search(ctx, keyword, page, size)
}

func (o *AdminDatabase) FindBlockUser(ctx context.Context, userIDs []string) ([]*table.ForbiddenAccount, error) {
	return o.forbiddenAccount.Find(ctx, userIDs)
}

func (o *AdminDatabase) SearchUserLimitLogin(ctx context.Context, keyword string, page int32, size int32) (uint32, []*table.LimitUserLoginIP, error) {
	return o.limitUserLoginIP.Search(ctx, keyword, page, size)
}

func (o *AdminDatabase) AddUserLimitLogin(ctx context.Context, ms []*table.LimitUserLoginIP) error {
	return o.limitUserLoginIP.Create(ctx, ms)
}

func (o *AdminDatabase) DelUserLimitLogin(ctx context.Context, ms []*table.LimitUserLoginIP) error {
	return o.limitUserLoginIP.Delete(ctx, ms)
}

func (o *AdminDatabase) CountLimitUserLoginIP(ctx context.Context, userID string) (uint32, error) {
	return o.limitUserLoginIP.Count(ctx, userID)
}

func (o *AdminDatabase) GetLimitUserLoginIP(ctx context.Context, userID string, ip string) (*table.LimitUserLoginIP, error) {
	return o.limitUserLoginIP.Take(ctx, userID, ip)
}

func (o *AdminDatabase) CacheToken(ctx context.Context, userID string, token string) error {
	return o.cache.AddTokenFlag(ctx, userID, token, constant.NormalToken)
}

func (o *AdminDatabase) GetTokens(ctx context.Context, userID string) (map[string]int32, error) {
	return o.cache.GetTokensWithoutError(ctx, userID)
}
