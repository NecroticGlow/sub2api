export default {
  intelligence: {
    title: 'Intelligence test',
    description: 'OpenAI OAuth accounts only. Uses gpt-6-astra and the v2.8.11 intelligence questions without browsing tools.',
    note: 'The candy answer 21 passes automatically; other answers require human review. Pelican HTML passes when it contains SVG. Results are stored on the server for all administrators.',
    search: 'Search account names', empty: 'No OpenAI OAuth accounts',
    test: 'Start test', testing: 'Testing…', idle: 'Not tested', question: 'Test question', candy: 'Candy (expected 21)', pelican: 'Pelican SVG HTML', knowledge: 'No-browse latest-info', reasoning: 'Reasoning effort',
    passed: 'Normal (reference matched)', manual_review: 'Needs human review', error: 'Test failed',
    answer: 'View raw answer and reference checks', review: 'Human review:', previous: 'Previous test', lastTest: 'Last test result', history: 'Test history', concurrency: 'Current concurrency', tokens: 'Total tokens', inputOutput: 'Input / output tokens', originalCost: 'Original cost',
    normal: 'Normal', degraded: 'Degraded', reset: 'Clear decision',
    reviewed_normal: 'Normal (human review)', reviewed_degraded: 'Degraded (human review)',
    items: { iphone: 'iPhone', announcement: 'Official announcement', on_sale: 'Official on-sale date', nvidia: 'NVIDIA GPU', android: 'Android', macos: 'macOS', windows: 'Windows', candy: 'Candy question', pelican: 'Pelican HTML' }
  }
}
