import express from 'express';
import { Readability } from '@mozilla/readability';
import { JSDOM } from 'jsdom';
import fetch from 'node-fetch';
import createDOMPurify from 'dompurify';

const app = express();
app.use(express.json());

app.post('/process', async (req, res) => {
    try {
        const { url } = req.body;
        if (!url) {
            return res.status(400).json({ error: 'URL is required' });
        }

        // Fetch the URL content
        const response = await fetch(url);
        const html = await response.text();


        // Create a new JSDOM instance
        const dom = new JSDOM(html, { url });

        // Use Readability to parse the document
        const reader = new Readability(dom.window.document);
        let article = reader.parse();

        if (!article) {
            return res.status(500).json({ error: 'Failed to parse the article' });
        }

        // Create a DOMPurify instance using the same JSDOM window
        const DOMPurify = createDOMPurify(dom.window);

        // Sanitize the article content
        const sanitizedContent = DOMPurify.sanitize(article.content);

        res.json({
            title: article.title,
            content: sanitizedContent,
            url: url
        });
    } catch (error) {
        console.error('Error processing URL:', error);
        res.status(500).json({ error: 'Failed to process URL' });
    }
});

const PORT = process.env.PORT || 3000;
app.listen(PORT, () => console.log(`Server running on port ${PORT}`));
