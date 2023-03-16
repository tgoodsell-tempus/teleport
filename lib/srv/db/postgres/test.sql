create or replace procedure teleport_create_user(username varchar, roles varchar[])
language plpgsql
as $$
declare
    role_ varchar;
begin
    if exists (select usename from pg_user , pg_group where pg_user.usesysid = any(pg_group.grolist) and pg_group.groname='teleport' and pg_user.usename=username) then
        execute format('alter user %I with login', username);
    else
        execute format('create user %I in role teleport', username);
    end if;
    foreach role_ in array roles
    loop
        execute format('grant %I to %I', role_, username);
    end loop;
end;$$;





create or replace procedure teleport_delete_%v()
language plpgsql
as $$
begin
    if exists (select usename from pg_stat_activity where usename = '%v') then
        raise notice 'User has active connections';
    else
        drop user %v;
    end if;
end;$$;