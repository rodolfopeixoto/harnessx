import { useEffect, useState } from "react";

type Status<T> =
  | { kind: "loading" }
  | { kind: "error"; error: unknown }
  | { kind: "ready"; data: T };

export function useFetched<T>(fn: () => Promise<T>, deps: unknown[] = []): Status<T> {
  const [s, setS] = useState<Status<T>>({ kind: "loading" });
  useEffect(() => {
    let cancelled = false;
    setS({ kind: "loading" });
    fn()
      .then((data) => !cancelled && setS({ kind: "ready", data }))
      .catch((error) => !cancelled && setS({ kind: "error", error }));
    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);
  return s;
}
