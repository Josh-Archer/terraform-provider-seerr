package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
)

func NewEmailSettingsDataSource() datasource.DataSource {
	return newNotificationClientDataSourceWithTypeName("email", "email_settings")
}
