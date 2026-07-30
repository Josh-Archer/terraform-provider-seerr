package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func NewPushbulletSettingsDataSource() datasource.DataSource {
	return newNotificationClientDataSourceWithTypeName("pushbullet", "pushbullet_settings")
}
