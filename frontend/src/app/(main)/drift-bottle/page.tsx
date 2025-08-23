'use client'

import { redirect } from 'next/navigation'

export default function DriftBottlePage() {
  redirect('/letters/write?type=drift')
}