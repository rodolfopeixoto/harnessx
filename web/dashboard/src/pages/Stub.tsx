import type { ReactNode } from "react";
import { Card, EmptyState } from "../ds";

type Props = {
  id: string;
  title: string;
  description: string;
  next?: ReactNode;
};

export function StubPage({ id, title, description, next }: Props) {
  return (
    <Card>
      <div data-testid={`page-${id}`}>
        <EmptyState
          title={title}
          hint={description}
          action={next}
        />
      </div>
    </Card>
  );
}
