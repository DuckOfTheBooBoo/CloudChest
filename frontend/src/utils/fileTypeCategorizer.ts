export type FileCategory = 'image' | 'video' | 'audio' | 'document' | 'archive' | 'font' | 'other' | 'plaintext';

export class FileTypeCategorizer {
  private static readonly categoryPatterns: Record<FileCategory, RegExp[]> = {
    image: [
      /^image\//,
      /\.(jpg|jpeg|png|gif|bmp|webp|svg|ico)$/i
    ],
    video: [
      /^video\//,
      /\.(mp4|webm|mov|avi|wmv|flv|mkv)$/i
    ],
    audio: [
      /^audio\//,
      /\.(mp3|wav|ogg|m4a|aac|flac)$/i
    ],
    document: [
      /^application\/(pdf|msword|vnd\.openxmlformats|vnd\.ms-|x-)/,
      /^text\/(plain|html|csv|markdown)/,
      /\.(pdf|doc|docx|xls|xlsx|ppt|pptx|txt|md|csv|html|rtf)$/i
    ],
    archive: [
      /^application\/(zip|x-zip|x-rar|x-7z|x-tar|x-gzip)/,
      /\.(zip|rar|7z|tar|gz|bz2)$/i
    ],
    font: [
      /^font\//,
      /^application\/font-/,
      /\.(ttf|woff|woff2|eot|otf)$/i
    ],
    plaintext: [
      /^text\//,
      /\.txt$/i
    ],
    other: [/.*/] // Catch-all pattern
  };

  private static readonly previewableCategories: Set<FileCategory> = new Set([
    'image',
    'video',
    'audio',
    'plaintext'
  ]);

  /**
   * Categorizes a file based on its MIME type and/or filename
   */
  static categorizeFile(mimeType: string, filename?: string): FileCategory {
    // Normalize MIME type to lowercase
    mimeType = mimeType.toLowerCase();

    // Try categorizing by MIME type first
    for (const [category, patterns] of Object.entries(FileTypeCategorizer.categoryPatterns)) {
      if (patterns.some(pattern => pattern.test(mimeType))) {
        return category as FileCategory;
      }
    }

    // If filename is provided and MIME type didn't give a specific match, try file extension
    if (filename) {
      for (const [category, patterns] of Object.entries(FileTypeCategorizer.categoryPatterns)) {
        if (patterns.some(pattern => pattern.test(filename))) {
          return category as FileCategory;
        }
      }
    }

    return 'other';
  }

  /**
   * Checks if a file is previewable based on its category
   */
  static isPreviewable(mimeType: string, filename?: string): boolean {
    const category = this.categorizeFile(mimeType, filename);
    return this.previewableCategories.has(category);
  }

  /**
   * Checks if a file might be an image regardless of its reported MIME type
   */
  static isProbablyImage(mimeType: string, filename?: string): boolean {
    return this.categoryPatterns.image.some(pattern => 
      pattern.test(mimeType) || (filename && pattern.test(filename))
    );
  }
}