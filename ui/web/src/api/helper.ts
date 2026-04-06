import type { Get, UnionToIntersection } from 'type-fest';

export type UrlPaths<T> = keyof T;

export type HttpMethods<T> = keyof UnionToIntersection<T[keyof T]>;

export type HttpMethodsFilteredByPath<
  T,
  Path extends UrlPaths<T>,
> = HttpMethods<T> & keyof UnionToIntersection<T[Path]>;

export type RequestParameters<
  T,
  Path extends string & UrlPaths<T>,
  Method extends HttpMethods<T> & string,
> = Get<T, `${Path}.${Method}.parameters.query`>;

export type RequestData<
  T,
  Path extends string & UrlPaths<T>,
  Method extends HttpMethods<T> & string,
> = Get<T, `${Path}.${Method}.requestBody.content.application/json`>;

export type ResponseData<
  T,
  Path extends string & UrlPaths<T>,
  Method extends HttpMethods<T> & string,
> = Get<T, `${Path}.${Method}.responses.200.content.application/json`>;
