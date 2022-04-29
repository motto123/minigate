-- auto-generated definition
create database minigate;
use minigate;
create table user
(
    id          int auto_increment
        primary key,
    account     char(30)                            not null,
    password    char(30)                            not null,
    nickname    varchar(20)                         null,
    create_time timestamp default CURRENT_TIMESTAMP null,
    update_time timestamp default CURRENT_TIMESTAMP null,
    constraint user_account_uindex
        unique (account),
    constraint user_password_uindex
        unique (password)
);

