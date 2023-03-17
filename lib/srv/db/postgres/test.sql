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



select r.rolname from pg_catalog.pg_auth_members m join pg_catalog.pg_roles r on r.oid = m.member where r.rolname = 'teleport';

select oid from pg_catalog.pg_roles where rolname = 'teleport';
select oid from pg_catalog.pg_roles where rolname = 'qweasdzxc';

select * from pg_catalog.pg_auth_members where roleid = (select oid from pg_catalog.pg_roles where rolname = 'teleport') and member = (select oid from pg_catalog.pg_roles where rolname = 'qweasdzxc')



create or replace procedure teleport_delete_user(username varchar)
language plpgsql
as $$
declare
    role_ varchar;
begin
    if exists (select usename from pg_stat_activity where usename = username) then
        raise notice 'User has active connections';
    else
    	for role_ in select a.rolname from pg_roles a where pg_has_role(username, a.oid, 'member') and a.rolname not in (username, 'teleport')
	    loop
        	execute format('revoke %I from %I', role_, username);
	    end loop;
        execute format('alter user %I with nologin', username);
    end if;
end;$$;