// Короткие алиасы на сгенерированные типы. Вручную модель данных НЕ описываем:
// единственный источник правды — api/openapi.yaml → npm run gen:api
import type { components } from './schema'

type S = components['schemas']

export type ID = S['ID']
export type Pt = S['Pt']
export type MapDoc = S['MapDoc']
export type MapSummary = S['MapSummary']
export type Floor = S['Floor']
export type Plan = S['Plan']
export type Room = S['Room']
export type Path = S['Path']
export type Poi = S['Poi']
export type Panorama = S['Panorama']
export type Hotspot = S['Hotspot']
export type PanoramaMedia = S['PanoramaMedia']
export type PublishResponse = S['PublishResponse']
export type User = S['User']
export type AuthResponse = S['AuthResponse']
