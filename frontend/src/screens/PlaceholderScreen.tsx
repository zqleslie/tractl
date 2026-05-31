import { Button } from '@/components/primitives'
import { useUiStore } from '@/stores/uiStore'

export function PlaceholderScreen({
  title,
  testId,
  subtitle,
}: {
  title: string
  testId: string
  subtitle?: string
}) {
  return (
    <div className="p-8" data-testid={testId}>
      <h1 className="text-ui-base font-medium text-text">{title}</h1>
      <p className="mt-2 text-ui-sm text-text-muted">
        {subtitle ??
          'Screen shell is ready. Workflow canvas and authoring panels come next.'}
      </p>
      <Button
        variant="ghost"
        className="mt-4"
        onClick={() => useUiStore.getState().setActiveScreen('fast-start')}
      >
        Back to Fast Start
      </Button>
    </div>
  )
}
