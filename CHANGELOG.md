# Change Log

## v0.3.0 (2022-03-23)

- Breaking API change: `Driver.PreGUI` has been dropped and `Driver.SetUIScale`
  has been renamed to `Driver.SetRenderScale`. The reason for this change is
  that setting the render scale in the middle of the frame had undesirable
  results.

## v0.2.1 (2022-03-23)

- Bug fix: Renderer scale was being "restored" to 0x0

## v0.2.0 (2022-03-23)

- Breaking API change: Dropped methods `Driver.PreRender` and
  `Driver.PostRender`. Frame lifecycle now has five phases:
    1. `FrameStart`
    2. Conventional rendering
    3. `PreGUI`
    4. GUI operations
    5. `FrameEnd`

## v0.1.1 (2022-03-23)

- Bug fix: Original clipping rectangle was not being restored at the end of
  `Driver.PostRender`

## v0.1.0 (2022-03-23)

- First release, including `Driver`, `EventHandler`, `NkDriver`, and `SDLDriver`
- Demo updated to use the library
- Dependencies updated to their latest versions
