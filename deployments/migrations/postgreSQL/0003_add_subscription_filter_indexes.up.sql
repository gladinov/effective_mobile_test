create index subscriptions_user_id_idx
    on subscriptions (user_id);

create index subscriptions_service_name_idx
    on subscriptions (service_name);

create index subscriptions_start_date_idx
    on subscriptions (start_date);

create index subscriptions_end_date_idx
    on subscriptions (end_date);
