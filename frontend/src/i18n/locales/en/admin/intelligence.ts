export default {
  intelligence: {
    title: 'Intelligence test',
    description: 'OpenAI OAuth accounts only. Uses gpt-6-astra and a fixed knowledge prompt without browsing tools.',
    note: 'Answers matching every reference item pass; other answers require human review. This is a screening heuristic, not a capability assessment. Each test uses upstream quota. Results and manual decisions are kept only while this page is open.',
    search: 'Search account names', empty: 'No OpenAI OAuth accounts',
    test: 'Start test', testing: 'Testing…', idle: 'Not tested',
    passed: 'Normal (reference matched)', manual_review: 'Needs human review', error: 'Test failed',
    answer: 'View raw answer and reference checks', review: 'Human review:', previous: 'Previous test', lastTest: 'Last test result', concurrency: 'Current concurrency', tokens: 'Total tokens', inputOutput: 'Input / output tokens', originalCost: 'Original cost',
    normal: 'Normal', degraded: 'Degraded', reset: 'Clear decision',
    reviewed_normal: 'Normal (human review)', reviewed_degraded: 'Degraded (human review)',
    items: { iphone: 'iPhone', announcement: 'Official announcement', on_sale: 'Official on-sale date', nvidia: 'NVIDIA GPU', android: 'Android', macos: 'macOS', windows: 'Windows' }
  }
}
