alter table subscriptions
    add constraint subscriptions_end_date_after_start_date_check
        check (end_date is null or end_date >= start_date);
