import { getCollection } from 'astro:content';

export async function GET() {
  const posts = await getCollection('blog');
  
  const results = posts.map(post => ({
    id: post.id,
    title: post.data.title,
    description: post.data.description || '',
    pubDate: post.data.pubDate,
  })).sort((a, b) => b.pubDate.valueOf() - a.pubDate.valueOf());

  return new Response(JSON.stringify(results), {
    headers: {
      'Content-Type': 'application/json'
    }
  });
}
