export function WebDesktopBanner() {
  return (
    <div
      className="border-b-[0.5px] border-warning bg-warning-bg px-4 py-2 text-ui-xs text-warning-fg"
      data-testid="web-desktop-banner"
      role="status"
    >
      Start traCtl Desktop to continue. The web surface connects to the local
      engine at localhost:7428.
    </div>
  )
}
