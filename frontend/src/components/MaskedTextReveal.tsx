import React from 'react'
import { motion, Variants } from 'framer-motion'

interface MaskedTextRevealProps {
  text: string
  className?: string
  highlightWords?: string[]
  highlightClass?: string
  delay?: number
}

export const MaskedTextReveal: React.FC<MaskedTextRevealProps> = ({
  text,
  className = '',
  highlightWords = [],
  highlightClass = 'text-transparent bg-clip-text bg-gradient-to-r from-brand-orange via-amber-500 to-rose-500',
  delay = 0,
}) => {
  const words = text.split(' ')

  const containerVariants: Variants = {
    hidden: {},
    visible: {
      transition: {
        staggerChildren: 0.04,
        delayChildren: delay,
      },
    },
  }

  const wordVariants: Variants = {
    hidden: {
      y: '115%',
      opacity: 0,
      rotateX: 20,
    },
    visible: {
      y: 0,
      opacity: 1,
      rotateX: 0,
      transition: {
        duration: 0.75,
        ease: [0.16, 1, 0.3, 1],
      },
    },
  }

  return (
    <motion.div
      variants={containerVariants}
      initial="hidden"
      animate="visible"
      className={`flex flex-wrap ${className}`}
    >
      {words.map((word, idx) => {
        const cleanWord = word.replace(/[^a-zA-Z0-9$/]/g, '')
        const isHighlight = highlightWords.some(
          (hw) => cleanWord.toLowerCase() === hw.toLowerCase()
        )

        return (
          <span key={idx} className="overflow-hidden inline-block mr-[0.25em] last:mr-0 py-1">
            <motion.span
              variants={wordVariants}
              className={`inline-block ${isHighlight ? highlightClass : ''}`}
            >
              {word}
            </motion.span>
          </span>
        )
      })}
    </motion.div>
  )
}
