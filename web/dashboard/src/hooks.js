import { useEffect, useState } from "react";
export function useFetched(fn, deps = []) {
    const [s, setS] = useState({ kind: "loading" });
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
