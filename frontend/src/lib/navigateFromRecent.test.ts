import { describe, expect, it } from 'vitest'
import { screenForRecentItem } from '@/lib/navigateFromRecent'

describe('screenForRecentItem', () => {
  it('routes requests to the request editor', () => {
    expect(
      screenForRecentItem({
        id: '1',
        name: 'GET /health',
        type: 'request',
        timeAgo: 'now',
      }),
    ).toBe('request')
  })

  it('routes workflows to the workflow placeholder', () => {
    expect(
      screenForRecentItem({
        id: '2',
        name: 'auth-flow',
        type: 'workflow',
        timeAgo: 'now',
      }),
    ).toBe('workflow')
  })
})
