Go Lock Cost
============

Channels in go make it easy to use a pattern of many workers drawing jobs from
a queue and dumping results into a queue. However, channels are not free. Under
the hood, traditional locking is required to produce the behavior of go's
channels.

The question: how much efficiency is lost as the use of channels increases?

Sub-question: how many channel operations per second is reasonable?

Investigation
-------------

To find out, I threw together the accompanying code that simulates multiple
workers doing work. A fixed amount of time ("work") is divided into jobs of
equal size. Each jobs is sent down a channel to a worker who sleeps for the
allotted time, and then sends the job into a result channel.

One receiver tallies the jobs as they come in and signals when all work is
done.

The smaller the size of each job, the more channel operations need to occur to
get all of the work done. Batching work into larger jobs should require less
locking and thus be more efficient, but how much?

Results
-------

The data below shows 5 different job sizes in a scenario of 3 workers doing
12 seconds of "work" Each block specifies a job size, how long the whole run
took, and a percentage measurement of how much extra time was spent not doing
work. In a perfect world, they would finish in 4 seconds with 0% overhead.

The test environment:

* go 1.1.2
* linux 3.11.0 on x86_64
* Intel Core2 Duo T8300 (not very recent)

::

    job size:  100us
    done in  9.044152192s
    overhead: 126.10%

    job size:  1ms
    done in  4.533251051s
    overhead: 13.33%

    job size:  10ms
    done in  4.052578229s
    overhead: 1.31%

    job size:  100ms
    done in  4.00572584s
    overhead: 0.14%

    job size:  1s
    done in  4.000667326s
    overhead: 0.02%

Analysis
--------

It is interesting to note that the overhead penalty is roughly linear, dropping
by a factor of 10 each time the job size increases by a factor of 10.

The first block shows a job size of 100 microseconds. That's 10,000 jobs to get
through one second of work! Each job requires a channel operation on the way
in, and another on the way out, so that's 20,000 channel operations for one
second of work. As you might expect, performance was terrible. It took 9
seconds to spend 4 seconds doing work, causing the whole run to take 126%
longer than the ideal. Ouch!

A job size of 1 millisecond is much more reasonable, but it still took 13%
longer to do the work.

With a job size of 10ms, we're only doing 100 jobs to complete one second of
work, and thus 200 channel operations. In this case, our penalty was only 1.3%,
which is approaching a reasonable cost for using channels.

Given that the penalty is linear, we can safely calculate from the first run
that since an extra 5 seconds were spent performing 240,000 channel operations,
each channel operation took about 21 microseconds. To state the obvious, that
adds up!

Variations
----------

Changing the number of workers did not have a large impact on the amount of
overhead.

On my CPU, which has 2 cores and no hyper-threading, setting MAXGOPROCS to 2
increased the overhead to 136%, 15%, 1.5%, 0.16%, and 0.02% respectively.

Conclusions
-----------

It is clear that when using channels to pass jobs to and from workers, it pays
to batch as much work as you reasonably can into each job.

As always, if performance is a concern, you should take your own measurements
and determine how much channel overhead is acceptable for your particular use
case. That said, this demonstration shows that channels should not be used to
feed workers without giving at least some thought to how much time a worker
will spend on each job.
