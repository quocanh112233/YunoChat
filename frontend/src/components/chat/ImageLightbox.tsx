'use client';

import React from 'react';
import { motion, AnimatePresence } from 'framer-motion';
import { X, Download } from 'lucide-react';

interface ImageLightboxProps {
  url: string | null;
  onClose: () => void;
}

export default function ImageLightbox({ url, onClose }: ImageLightboxProps) {
  if (!url) return null;

  return (
    <AnimatePresence>
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="fixed inset-0 z-100 flex items-center justify-center bg-black/95 backdrop-blur-sm p-4 md:p-10"
        onClick={onClose}
      >
        <motion.button
          initial={{ y: -20, opacity: 0 }}
          animate={{ y: 0, opacity: 1 }}
          className="absolute top-5 right-5 p-2 bg-slate-800/50 hover:bg-slate-700/50 text-white rounded-full transition-colors z-110"
          onClick={(e) => {
            e.stopPropagation();
            onClose();
          }}
        >
          <X className="w-6 h-6" />
        </motion.button>

        <div className="absolute top-5 left-5 flex gap-2 z-110">
          <a
            href={url}
            download
            target="_blank"
            rel="noopener noreferrer"
            className="p-2 bg-slate-800/50 hover:bg-slate-700/50 text-white rounded-full transition-colors"
            onClick={(e) => e.stopPropagation()}
          >
            <Download className="w-5 h-5" />
          </a>
        </div>

        <motion.div
          initial={{ scale: 0.9, opacity: 0 }}
          animate={{ scale: 1, opacity: 1 }}
          exit={{ scale: 0.9, opacity: 0 }}
          transition={{ type: 'spring', damping: 25, stiffness: 300 }}
          className="relative max-w-full max-h-full flex items-center justify-center"
          onClick={(e) => e.stopPropagation()}
        >
          <img
            src={url}
            alt="Full size"
            className="max-w-full max-h-[90vh] object-contain shadow-2xl rounded-sm selection:bg-transparent"
            draggable={false}
          />
        </motion.div>
        
        <div className="absolute bottom-5 left-1/2 -translate-x-1/2 px-4 py-2 bg-slate-800/40 backdrop-blur-md rounded-full border border-slate-700/50 pointer-events-none">
          <p className="text-xs text-slate-300 font-medium">Click outside to close</p>
        </div>
      </motion.div>
    </AnimatePresence>
  );
}
