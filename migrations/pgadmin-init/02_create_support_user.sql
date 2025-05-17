-- Создаем роль support_user
CREATE ROLE support_user LOGIN PASSWORD 'ThatParrooll123QWEXZC!';
GRANT CONNECT ON DATABASE edusync TO support_user;
GRANT USAGE ON SCHEMA public TO support_user;

GRANT SELECT ON ALL TABLES IN SCHEMA public TO support_user;
REVOKE UPDATE, DELETE ON public.users FROM support_user;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT ON TABLES TO support_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA public
  REVOKE INSERT, UPDATE, DELETE ON TABLES FROM support_user;


REVOKE UPDATE, DELETE ON public.users FROM support_user;
