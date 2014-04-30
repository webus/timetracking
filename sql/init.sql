-- DROP TABLE currency;
CREATE TABLE currency
(
  uid bigserial NOT NULL,
  currency_name character varying(100) NOT NULL,
  CONSTRAINT currency_pkey PRIMARY KEY (uid),
  CONSTRAINT currency_currency_name_key UNIQUE (currency_name)
);

-- DROP TABLE projects;
CREATE TABLE projects
(
  uid bigserial NOT NULL,
  project_name character varying(1000) NOT NULL,
  CONSTRAINT projects_pkey PRIMARY KEY (uid),
  CONSTRAINT projects_project_name_key UNIQUE (project_name)
);

-- DROP TABLE notes;
CREATE TABLE notes
(
  uid bigserial NOT NULL,
  project_id bigint NOT NULL,
  note_text text,
  date_add timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT notes_pkey PRIMARY KEY (uid),
  CONSTRAINT notes_project_id_fkey FOREIGN KEY (project_id)
      REFERENCES projects (uid) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- DROP TABLE rates;
CREATE TABLE rates
(
  uid bigserial NOT NULL,
  project_id bigint NOT NULL,
  rate numeric(15,2) NOT NULL,
  currency_id bigint NOT NULL,
  from_date timestamp without time zone NOT NULL,
  to_date timestamp without time zone NOT NULL,
  CONSTRAINT rates_pkey PRIMARY KEY (uid),
  CONSTRAINT rates_currency_id_fkey FOREIGN KEY (currency_id)
      REFERENCES currency (uid) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT rates_project_id_fkey FOREIGN KEY (project_id)
      REFERENCES projects (uid) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- DROP TABLE states;
CREATE TABLE states
(
  uid bigserial NOT NULL,
  state_name character varying(100) NOT NULL,
  CONSTRAINT states_pkey PRIMARY KEY (uid),
  CONSTRAINT states_state_name_key UNIQUE (state_name)
);

-- DROP TABLE workday;
CREATE TABLE workday
(
  uid bigserial NOT NULL,
  state_id bigint NOT NULL,
  action_time timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT workday_pkey PRIMARY KEY (uid),
  CONSTRAINT workday_state_id_fkey FOREIGN KEY (state_id)
      REFERENCES states (uid) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- DROP TABLE worklog;
CREATE TABLE worklog
(
  uid bigserial NOT NULL,
  project_id bigint NOT NULL,
  state_id bigint NOT NULL,
  action_time timestamp with time zone NOT NULL DEFAULT now(),
  action_comment character varying(1000),
  CONSTRAINT worklog_pkey PRIMARY KEY (uid),
  CONSTRAINT worklog_project_id_fkey FOREIGN KEY (project_id)
      REFERENCES projects (uid) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE NO ACTION,
  CONSTRAINT worklog_state_id_fkey FOREIGN KEY (state_id)
      REFERENCES states (uid) MATCH SIMPLE
      ON UPDATE NO ACTION ON DELETE NO ACTION
);

-- DROP FUNCTION start_project(character varying);
CREATE OR REPLACE FUNCTION start_project(project_name character varying)
  RETURNS bigint AS
$BODY$
declare
        ttt varchar;
begin
        select st.state_name into ttt from worklog wl
        inner join projects pr on pr.uid = wl.project_id
        inner join states st on st.uid = wl.state_id
        where pr.project_name = project_name order by wl.action_time desc limit 1;
end;
$BODY$
  LANGUAGE plpgsql;

-- DROP FUNCTION start_project(character varying, character varying);
CREATE OR REPLACE FUNCTION start_project(_project_name character varying, _project_comment character varying)
  RETURNS bigint AS
$BODY$
declare
        ttt varchar;
begin
        select st.state_name into ttt from worklog wl
        inner join projects pr on pr.uid = wl.project_id
        inner join states st on st.uid = wl.state_id
        where pr.project_name = _project_name order by wl.action_time desc limit 1;
        if ttt = 'start' then
                return 1;
        else
                insert into worklog(project_id,state_id,action_comment)
                values(
                        (select uid from projects where project_name = _project_name),
                        (select uid from states where state_name = 'start'),
                        _project_comment
                );
                return 0;
        end if;
end;
$BODY$
  LANGUAGE plpgsql;

-- DROP FUNCTION start_workday();
CREATE OR REPLACE FUNCTION start_workday()
  RETURNS bigint AS
$BODY$
declare
        ttt varchar;
begin
        select st.state_name into ttt from workday wd
        inner join states st on st.uid = wd.state_id
        order by wd.action_time desc limit 1;
        if ttt = 'start' then
                return 1;
        else
                insert into workday(state_id)
                values(
                        (select uid from states where state_name = 'start')
                );
                return 0;
        end if;
end;
$BODY$
  LANGUAGE plpgsql;

-- DROP FUNCTION stop_project(character varying, character varying);
CREATE OR REPLACE FUNCTION stop_project(_project_name character varying, _project_comment character varying)
  RETURNS bigint AS
$BODY$
declare
        ttt varchar;
begin
        select st.state_name into ttt from worklog wl
        inner join projects pr on pr.uid = wl.project_id
        inner join states st on st.uid = wl.state_id
        where pr.project_name = _project_name order by wl.action_time desc limit 1;
        if ttt = 'stop' then
                return 1;
        else
                insert into worklog(project_id,state_id,action_comment)
                values(
                        (select uid from projects where project_name = _project_name),
                        (select uid from states where state_name = 'stop'),
                        _project_comment
                );
                return 0;
        end if;
end;
$BODY$
  LANGUAGE plpgsql;

-- DROP FUNCTION stop_workday();
CREATE OR REPLACE FUNCTION stop_workday()
  RETURNS bigint AS
$BODY$
declare
        ttt varchar;
begin
        select st.state_name into ttt from workday wd
        inner join states st on st.uid = wd.state_id
        order by wd.action_time desc limit 1;
        if ttt = 'stop' then
                return 1;
        else
                insert into workday(state_id)
                values(
                        (select uid from states where state_name = 'stop')
                );
                return 0;
        end if;
end;
$BODY$
  LANGUAGE plpgsql;
