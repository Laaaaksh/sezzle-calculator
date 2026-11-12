# Accessibility

This is the result of the accessibility pass. Each section names the check, what changed and why. Every check a machine runs is a test, so a regression fails the build.

## Contrast

- Check: text needs 4.5 to 1, interface parts (focus ring, active operator ring) need 3 to 1, in both themes.
- Change: `src/styles/contrast.test.ts` reads the colours from `src/styles/tokens.css` and computes the WCAG ratio for 25 text and interface pairs in each theme. The smallest text ratio is 4.7 (operator label on a hovered key, light theme). The smallest interface ratio is 5.3 (active fill next to a tinted key, light theme).
- Why: axe cannot check contrast in jsdom, because jsdom has no layout and no real styles. The axe helper turns off only that rule and says so in a comment. The token test covers the gap.
- Decision: key borders use the brief's `#E5E7EB`, which is about 1.2 to 1 against a white card. Every key carries a text label, so the border is not what identifies it. WCAG 1.4.11 does not ask for 3 to 1 on a border in that case.

## Focus ring

- Check: every interactive element shows a visible focus ring.
- Change: all controls use a 2px outline in `--color-focus`, set to the text colour so it has the highest contrast on each surface and cannot be mistaken for the accent ring. `stylesheets.test.ts` fails if any CSS module removes an outline, or has a clickable rule (`cursor: pointer`) without a `:focus-visible` rule of at least 2px.
- Fix found: the focus outline sat on top of the active operator ring and hid it. A focused active operator now moves its outline out by 5px, so both show.

## Tab order

- Check: Tab visits controls in reading order.
- Change: none needed. The page order is theme options, keys in keypad order, then the History button. `App.test.tsx` tabs through the page and checks each stop, and checks that no element uses a positive `tabindex`.
- Note: the History panel hides with `display: none` when closed on narrow screens, so its entries stay out of the tab order until you open it.

## Touch targets

- Check: every control is at least 44px tall and wide.
- Change: `--size-touch` is 2.75rem (44px) and keys use `--size-key` at 3.5rem. The theme options, the History button, the history entries and the Retry button all use these. `stylesheets.test.ts` fails if a clickable rule has no `min-height` from those tokens. Keys stay 57px wide at a 320px screen.
- Fix found: the Retry button and the history entries were under 44px tall. Both now meet it.

## Active operator

- Check: the active operator is not shown by colour alone.
- Change: the key gets an accent fill, a two-part ring (a gap, then a 2px accent line) and `aria-pressed="true"`, so screen readers announce it too. In Windows high contrast mode `box-shadow` is dropped, so the ring is repeated as an outline there. Tests check the ring rule, the forced-colors rule and the `aria-pressed` values.
- Fix found: the ring would have vanished in forced-colors mode, so the outline fallback was added.

## Errors

- Check: errors are announced.
- Change: the error message is now a `role="alert"` paragraph outside the polite status region. Before, it sat inside the polite region and queued behind the result and expression updates. Alerts are read at once. The Retry button stays outside the alert so only the message is spoken. Tests check the alert role and that the status region does not contain it.

## Zoom and reflow

- Check: zoom is not blocked and the layout holds at 200 percent.
- Change: `index.html` keeps `width=device-width, initial-scale=1.0` with no `user-scalable` or `maximum-scale`. `index.html.test.ts` fails if either is added. Sizes use rem, so text zoom scales the layout. The grid collapses to one column below 900px, so 200 percent zoom on a laptop gives the narrow layout. Long results shrink with the card width and wrap as a last resort, so nothing scrolls sideways down to 320px.
- Not automated: layout at zoom needs a real browser. The browser check (prompt 23 in [PROMPTS.md](PROMPTS.md)) covers it.

## axe in the component tests

- `src/test/axe.ts` runs axe-core on a rendered container and fails with the rule ids and elements.
- It runs on the calculator when idle, with a result, with an active operator, with a network error and Retry, with a divide by zero error, and with the history drawer open. It also runs on the theme toggle and on the whole app.
- `src/test/axe.test.tsx` proves the helper fails on a button with no name.

## Other guards added

- `stylesheets.test.ts` also checks that animation and transitions only exist inside `prefers-reduced-motion: no-preference`, and that no CSS module hard-codes a duration.
- `tokens.test.ts` checks that the system dark theme and the explicit dark theme hold identical values.
