-- +goose Up
-- +goose StatementBegin
CREATETABLEIFNOTEXISTS"config_roles"(
    permissionINTEGERNOTNULL,
    role_idTEXTNOTNULLPRIMARYKEY(permission, role_id)
) STRICT;
CREATETABLEIFNOTEXISTS"config_channels"(
purposeINTEGERNOTNULL,
channel_idTEXTNOTNULLPRIMARYKEY(purpose, channel_id)
) STRICT;
CREATETABLEIFNOTEXISTS"config_messages"(
purposeINTEGERPRIMARYKEY,
message_idTEXTNOTNULL,
channel_idTEXTNOTNULL,
UNIQUE(message_id, channel_id)
) STRICT;
-- +goose StatementEnd
-- +goose Down
-- +goose StatementBegin
DROPTABLEIFEXISTS"config_roles";
DROPTABLEIFEXISTS"config_channels";
DROPTABLEIFEXISTS"config_messages";
-- +goose StatementEnd
