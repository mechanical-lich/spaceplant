# Shaders

The game supports an optional post-process shader applied to the **map/tile render only**. The HUD and UI are drawn on top unaffected.

## How it works

1. `Level.ShaderSrc []byte` holds the raw KAGE shader source. It is loaded from `data/shaders/crt.kage` in `NewSimWorld`.
2. On the first call to `Level.Render`, the shader is compiled via `ebiten.NewShader` (lazy so it runs after the Ebiten graphics context is ready) and cached in `Level.shader`.
3. At the end of `Render`, if a shader is compiled, the output image is blitted through `DrawRectShader` into a new image which is returned instead.
4. Intensity is controlled by the `Intensity` uniform (a `float` in the KAGE source). It is read from `config.Global().CRTIntensity` each frame, so runtime changes take effect immediately.

## Adding a new shader

- Drop a `.kage` file into `data/shaders/`.
- Load the bytes and assign them to `sim.Level.ShaderSrc`. The existing `CRTIntensity` config key drives blending; add a new uniform and config key if the shader needs different parameters.
- The shader receives the rendered tile image as `imageSrc0`. Use `imageSrc0At(src)` to sample the original pixel.

## Shader file: `data/shaders/crt.kage`

A Kage port of the [shadertoy CRT shader by inigo quilez](https://www.shadertoy.com/view/Ms23DR) (CC BY-NC-SA 3.0). Effects include:

- Screen curvature warp
- Chromatic aberration (RGB channel offset)
- Vignette
- Scanline dimming
- Per-column pixel darkening

The `Intensity` uniform lerps between the original image and the full effect: `mix(original, processed, Intensity)`.

## Preserving the shader across level changes

When `RegenerateLevel` or a save load replaces `sim.Level`, `ShaderSrc` is copied to the new level before the pointer swap. If you add more shader-related fields to `Level`, copy them in both `RegenerateLevel` (`sim_world.go`) and the load path (`save.go`).
