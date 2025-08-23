'use client'

import { redirect } from 'next/navigation'

export default function FutureLetterPage() {
  redirect('/letters/write?type=future')
}