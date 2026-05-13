import {
  Message,
  DescMessage,
  MessageShape,
  DescService,
  DescMethodUnary,
  DescMethodServerStreaming,
  DescMethodClientStreaming,
  DescMethodBiDiStreaming,
} from "@bufbuild/protobuf";
import type { Transport, CallOptions } from "@connectrpc/connect";
import { createClient, type Client } from "@connectrpc/connect";

type Request<I extends DescMessage> = OmitTypeName<MessageShape<I>>;

// prettier-ignore
export type StrictClient<T extends DescService> = {
  [P in keyof T["method"]]:
  T["method"][P] extends DescMethodUnary<infer I, infer O> ? <R extends Request<I>>(request: R, options?: CallOptions) => Promise<MessageShape<O>>
  : T["method"][P] extends DescMethodServerStreaming<infer I, infer O> ? <R extends Request<I>>(request: R, options?: CallOptions) => AsyncIterable<MessageShape<O>>
  : T["method"][P] extends DescMethodClientStreaming<infer I, infer O> ? <R extends Request<I>>(request: AsyncIterable<R>, options?: CallOptions) => Promise<MessageShape<O>>
  : T["method"][P] extends DescMethodBiDiStreaming<infer I, infer O> ? <R extends Request<I>>(request: AsyncIterable<R>, options?: CallOptions) => AsyncIterable<MessageShape<O>>
  : never;
};

/**
 * Create a strict client from a service definition and a transport.
 *
 * With a strict client, all non-optional properties of the request message must
 * be specified.
 */
export function createStrictClient<S extends DescService>(
  service: S,
  transport: Transport,
) {
  return createClient(service, transport) as unknown as StrictClient<S>;
}

/**
 * Removes the $typeName and $unknown properties from a message.
 */
export type OmitTypeName<T extends Message> = {
  [P in keyof T as P extends "$typeName" | "$unknown" ? never : P]: Recurse<
    T[P]
  >;
};

// prettier-ignore
type Recurse<F> =
  F extends Array<infer U> ? Array<Recurse<U>>
  : F extends Message ? OmitTypeName<F>
  : F extends { case: infer C extends string; value: infer V extends Message } ? { case: C; value: OmitTypeName<V> }
  : F extends { [key: string]: infer V extends Message } ? { [key: string]: OmitTypeName<V> }
  : F extends { [key: number]: infer V extends Message } ? { [key: number]: OmitTypeName<V> }
  : F;
