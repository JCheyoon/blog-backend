-- RLS 활성화 전 상태로 되돌리기

DROP POLICY IF EXISTS "Enable read access for all users" ON public.categories;
ALTER TABLE public.categories DISABLE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS "Enable read access for all users" ON public.posts;
ALTER TABLE public.posts DISABLE ROW LEVEL SECURITY;

ALTER TABLE public.schema_migrations DISABLE ROW LEVEL SECURITY;
